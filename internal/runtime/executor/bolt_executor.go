package executor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/misc"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/util"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v6/sdk/translator"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// BoltExecutor implements streaming code generation and WebContainer execution context
// based on the Bolt.new open source implementation using Anthropic Claude API.
type BoltExecutor struct {
	cfg *config.Config
}

func NewBoltExecutor(cfg *config.Config) *BoltExecutor {
	return &BoltExecutor{cfg: cfg}
}

func (e *BoltExecutor) Identifier() string {
	return "bolt"
}

// PrepareRequest injects Bolt/Claude credentials into the outgoing HTTP request.
func (e *BoltExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	if req == nil {
		return nil
	}
	apiKey, _ := boltCreds(auth)
	if strings.TrimSpace(apiKey) == "" {
		return nil
	}

	// Use x-api-key header for Anthropic API
	req.Header.Del("Authorization")
	req.Header.Set("x-api-key", apiKey)

	var attrs map[string]string
	if auth != nil {
		attrs = auth.Attributes
	}
	util.ApplyCustomHeadersFromAttrs(req, attrs)
	return nil
}

// HttpRequest injects credentials and executes the request.
func (e *BoltExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("bolt executor: request is nil")
	}
	if ctx == nil {
		ctx = req.Context()
	}
	httpReq := req.WithContext(ctx)
	if err := e.PrepareRequest(httpReq, auth); err != nil {
		return nil, err
	}
	httpClient := newProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	return httpClient.Do(httpReq)
}

// Execute handles non-streaming requests with Bolt-specific context injection.
func (e *BoltExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (resp cliproxyexecutor.Response, err error) {
	apiKey, baseURL := boltCreds(auth)
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	reporter := newUsageReporter(ctx, e.Identifier(), req.Model, auth)
	defer reporter.trackFailure(ctx, &err)

	model := req.Model
	if override := e.resolveUpstreamModel(req.Model, auth); override != "" {
		model = override
	}

	from := opts.SourceFormat
	to := sdktranslator.FromString("claude")
	stream := from != to

	originalPayload := bytes.Clone(req.Payload)
	if len(opts.OriginalRequest) > 0 {
		originalPayload = bytes.Clone(opts.OriginalRequest)
	}

	originalTranslated := sdktranslator.TranslateRequest(from, to, model, originalPayload, stream)
	body := sdktranslator.TranslateRequest(from, to, model, bytes.Clone(req.Payload), stream)
	body, _ = sjson.SetBytes(body, "model", model)

	// Inject Bolt.new specific WebContainer execution context
	body = e.injectBoltContext(body, req.Metadata)

	// Apply thinking config for code generation
	body = e.injectThinkingConfig(model, req.Metadata, body)

	body = applyPayloadConfigWithRoot(e.cfg, model, to.String(), "", body, originalTranslated)
	body = ensureMaxTokensForThinking(model, body)

	var extraBetas []string
	extraBetas, body = extractAndRemoveBetas(body)

	url := fmt.Sprintf("%s/v1/messages", baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return resp, err
	}

	e.applyBoltHeaders(httpReq, auth, apiKey, false, extraBetas)

	var authID, authLabel, authType, authValue string
	if auth != nil {
		authID = auth.ID
		authLabel = auth.Label
		authType, authValue = auth.AccountInfo()
	}

	recordAPIRequest(ctx, e.cfg, upstreamRequestLog{
		URL:       url,
		Method:    http.MethodPost,
		Headers:   httpReq.Header.Clone(),
		Body:      body,
		Provider:  e.Identifier(),
		AuthID:    authID,
		AuthLabel: authLabel,
		AuthType:  authType,
		AuthValue: authValue,
	})

	httpClient := newProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		return resp, err
	}

	recordAPIResponseMetadata(ctx, e.cfg, httpResp.StatusCode, httpResp.Header.Clone())

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		b, _ := io.ReadAll(httpResp.Body)
		appendAPIResponseChunk(ctx, e.cfg, b)
		log.Debugf("request error, error status: %d, error body: %s", httpResp.StatusCode, summarizeErrorBody(httpResp.Header.Get("Content-Type"), b))
		err = statusErr{code: httpResp.StatusCode, msg: string(b)}
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
		return resp, err
	}

	decodedBody, err := decodeResponseBody(httpResp.Body, httpResp.Header.Get("Content-Encoding"))
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
		return resp, err
	}
	defer func() {
		if errClose := decodedBody.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
	}()

	data, err := io.ReadAll(decodedBody)
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		return resp, err
	}

	appendAPIResponseChunk(ctx, e.cfg, data)

	if stream {
		lines := bytes.Split(data, []byte("\n"))
		for _, line := range lines {
			if detail, ok := parseClaudeStreamUsage(line); ok {
				reporter.publish(ctx, detail)
			}
		}
	} else {
		reporter.publish(ctx, parseClaudeUsage(data))
	}

	var param any
	out := sdktranslator.TranslateNonStream(
		ctx,
		to,
		from,
		req.Model,
		bytes.Clone(opts.OriginalRequest),
		body,
		data,
		&param,
	)

	resp = cliproxyexecutor.Response{Payload: []byte(out)}
	return resp, nil
}

// ExecuteStream handles streaming code generation with WebContainer context.
func (e *BoltExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (stream <-chan cliproxyexecutor.StreamChunk, err error) {
	apiKey, baseURL := boltCreds(auth)
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	reporter := newUsageReporter(ctx, e.Identifier(), req.Model, auth)
	defer reporter.trackFailure(ctx, &err)

	from := opts.SourceFormat
	to := sdktranslator.FromString("claude")
	model := req.Model
	if override := e.resolveUpstreamModel(req.Model, auth); override != "" {
		model = override
	}

	originalPayload := bytes.Clone(req.Payload)
	if len(opts.OriginalRequest) > 0 {
		originalPayload = bytes.Clone(opts.OriginalRequest)
	}

	originalTranslated := sdktranslator.TranslateRequest(from, to, model, originalPayload, true)
	body := sdktranslator.TranslateRequest(from, to, model, bytes.Clone(req.Payload), true)
	body, _ = sjson.SetBytes(body, "model", model)

	// Inject Bolt.new specific WebContainer execution context
	body = e.injectBoltContext(body, req.Metadata)

	// Inject thinking config for streaming code generation
	body = e.injectThinkingConfig(model, req.Metadata, body)

	body = applyPayloadConfigWithRoot(e.cfg, model, to.String(), "", body, originalTranslated)
	body = ensureMaxTokensForThinking(model, body)

	var extraBetas []string
	extraBetas, body = extractAndRemoveBetas(body)

	url := fmt.Sprintf("%s/v1/messages", baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	e.applyBoltHeaders(httpReq, auth, apiKey, true, extraBetas)

	var authID, authLabel, authType, authValue string
	if auth != nil {
		authID = auth.ID
		authLabel = auth.Label
		authType, authValue = auth.AccountInfo()
	}

	recordAPIRequest(ctx, e.cfg, upstreamRequestLog{
		URL:       url,
		Method:    http.MethodPost,
		Headers:   httpReq.Header.Clone(),
		Body:      body,
		Provider:  e.Identifier(),
		AuthID:    authID,
		AuthLabel: authLabel,
		AuthType:  authType,
		AuthValue: authValue,
	})

	httpClient := newProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		return nil, err
	}

	recordAPIResponseMetadata(ctx, e.cfg, httpResp.StatusCode, httpResp.Header.Clone())

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		b, _ := io.ReadAll(httpResp.Body)
		appendAPIResponseChunk(ctx, e.cfg, b)
		log.Debugf("request error, error status: %d, error body: %s", httpResp.StatusCode, summarizeErrorBody(httpResp.Header.Get("Content-Type"), b))
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
		err = statusErr{code: httpResp.StatusCode, msg: string(b)}
		return nil, err
	}

	decodedBody, err := decodeResponseBody(httpResp.Body, httpResp.Header.Get("Content-Encoding"))
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
		return nil, err
	}

	out := make(chan cliproxyexecutor.StreamChunk)
	stream = out

	go func() {
		defer close(out)
		defer func() {
			if errClose := decodedBody.Close(); errClose != nil {
				log.Errorf("response body close error: %v", errClose)
			}
		}()

		// Direct SSE forwarding for Claude -> Claude (Bolt format)
		if from == to {
			scanner := bufio.NewScanner(decodedBody)
			scanner.Buffer(nil, 52_428_800) // 50MB
			for scanner.Scan() {
				line := scanner.Bytes()
				appendAPIResponseChunk(ctx, e.cfg, line)
				if detail, ok := parseClaudeStreamUsage(line); ok {
					reporter.publish(ctx, detail)
				}

				cloned := make([]byte, len(line)+1)
				copy(cloned, line)
				cloned[len(line)] = '\n'
				out <- cliproxyexecutor.StreamChunk{Payload: cloned}
			}
			if errScan := scanner.Err(); errScan != nil {
				recordAPIResponseError(ctx, e.cfg, errScan)
				reporter.publishFailure(ctx)
				out <- cliproxyexecutor.StreamChunk{Err: errScan}
			}
			return
		}

		// Translation for other formats
		scanner := bufio.NewScanner(decodedBody)
		scanner.Buffer(nil, 52_428_800) // 50MB
		var param any
		for scanner.Scan() {
			line := scanner.Bytes()
			appendAPIResponseChunk(ctx, e.cfg, line)
			if detail, ok := parseClaudeStreamUsage(line); ok {
				reporter.publish(ctx, detail)
			}

			chunks := sdktranslator.TranslateStream(
				ctx,
				to,
				from,
				req.Model,
				bytes.Clone(opts.OriginalRequest),
				body,
				bytes.Clone(line),
				&param,
			)
			for i := range chunks {
				out <- cliproxyexecutor.StreamChunk{Payload: []byte(chunks[i])}
			}
		}
		if errScan := scanner.Err(); errScan != nil {
			recordAPIResponseError(ctx, e.cfg, errScan)
			reporter.publishFailure(ctx)
			out <- cliproxyexecutor.StreamChunk{Err: errScan}
		}
	}()

	return stream, nil
}

// CountTokens returns token count for Bolt requests.
func (e *BoltExecutor) CountTokens(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	apiKey, baseURL := boltCreds(auth)
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	from := opts.SourceFormat
	to := sdktranslator.FromString("claude")
	stream := from != to

	model := req.Model
	if override := e.resolveUpstreamModel(req.Model, auth); override != "" {
		model = override
	}

	body := sdktranslator.TranslateRequest(from, to, model, bytes.Clone(req.Payload), stream)
	body, _ = sjson.SetBytes(body, "model", model)
	body = e.injectBoltContext(body, req.Metadata)

	var extraBetas []string
	extraBetas, body = extractAndRemoveBetas(body)

	url := fmt.Sprintf("%s/v1/messages/count_tokens", baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return cliproxyexecutor.Response{}, err
	}

	e.applyBoltHeaders(httpReq, auth, apiKey, false, extraBetas)

	var authID, authLabel, authType, authValue string
	if auth != nil {
		authID = auth.ID
		authLabel = auth.Label
		authType, authValue = auth.AccountInfo()
	}

	recordAPIRequest(ctx, e.cfg, upstreamRequestLog{
		URL:       url,
		Method:    http.MethodPost,
		Headers:   httpReq.Header.Clone(),
		Body:      body,
		Provider:  e.Identifier(),
		AuthID:    authID,
		AuthLabel: authLabel,
		AuthType:  authType,
		AuthValue: authValue,
	})

	httpClient := newProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		return cliproxyexecutor.Response{}, err
	}

	recordAPIResponseMetadata(ctx, e.cfg, resp.StatusCode, resp.Header.Clone())

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		appendAPIResponseChunk(ctx, e.cfg, b)
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
		return cliproxyexecutor.Response{}, statusErr{code: resp.StatusCode, msg: string(b)}
	}

	decodedBody, err := decodeResponseBody(resp.Body, resp.Header.Get("Content-Encoding"))
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
		return cliproxyexecutor.Response{}, err
	}
	defer func() {
		if errClose := decodedBody.Close(); errClose != nil {
			log.Errorf("response body close error: %v", errClose)
		}
	}()

	data, err := io.ReadAll(decodedBody)
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		return cliproxyexecutor.Response{}, err
	}

	appendAPIResponseChunk(ctx, e.cfg, data)

	count := gjson.GetBytes(data, "input_tokens").Int()
	out := sdktranslator.TranslateTokenCount(ctx, to, from, count, data)
	return cliproxyexecutor.Response{Payload: []byte(out)}, nil
}

// Refresh attempts to refresh Bolt credentials (currently no-op as Bolt uses API keys).
func (e *BoltExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	log.Debugf("bolt executor: refresh called")
	if auth == nil {
		return nil, fmt.Errorf("bolt executor: auth is nil")
	}
	// Bolt uses API keys, no refresh needed
	return auth, nil
}

// injectBoltContext adds Bolt.new specific WebContainer execution context to the system prompt.
func (e *BoltExecutor) injectBoltContext(body []byte, metadata map[string]any) []byte {
	// Check if metadata contains webcontainer flag
	if metadata != nil {
		if webContainer, ok := metadata["webcontainer"].(bool); ok && webContainer {
			// Inject WebContainer context into system prompt
			boltSystemContext := map[string]any{
				"type": "text",
				"text": "You are Bolt, an expert AI assistant and exceptional senior software developer with vast knowledge across multiple programming languages, frameworks, and best practices. Your capabilities include:\n\n" +
					"<bolt_capabilities>\n" +
					"* Creating and managing project structures\n" +
					"* Writing clean, efficient, and well-documented code\n" +
					"* Debugging complex issues and providing detailed explanations\n" +
					"* Offering architectural insights and design patterns\n" +
					"* Staying up-to-date with the latest technologies and best practices\n" +
					"* Reading and analyzing existing files in the project\n" +
					"* Listing files and directories to understand the project structure\n" +
					"* Performing web searches for additional information when needed\n" +
					"</bolt_capabilities>\n\n" +
					"<webcontainer_environment>\n" +
					"You are running in WebContainer, an in-browser Node.js runtime. Key characteristics:\n" +
					"* Commands run inside a Node.js environment with limited shell capabilities\n" +
					"* Filesystem is in-memory and browser-based\n" +
					"* Network requests are proxied through the browser\n" +
					"* You can install npm packages and run Node.js scripts\n" +
					"* Development servers can be started and will be accessible via browser preview\n" +
					"</webcontainer_environment>",
			}

			// Get existing system prompt
			system := gjson.GetBytes(body, "system")
			if system.Exists() && system.IsArray() {
				// Prepend Bolt context to existing system array
				systemArray := []any{boltSystemContext}
				system.ForEach(func(_, value gjson.Result) bool {
					var item any
					if err := gjson.Unmarshal([]byte(value.Raw), &item); err == nil {
						systemArray = append(systemArray, item)
					}
					return true
				})
				body, _ = sjson.SetBytes(body, "system", systemArray)
			} else {
				// Create new system array with Bolt context
				body, _ = sjson.SetBytes(body, "system", []any{boltSystemContext})
			}
		}
	}

	return body
}

// injectThinkingConfig adds thinking configuration for code generation tasks.
func (e *BoltExecutor) injectThinkingConfig(modelName string, metadata map[string]any, body []byte) []byte {
	budget, ok := util.ResolveClaudeThinkingConfig(modelName, metadata)
	if !ok {
		return body
	}
	return util.ApplyClaudeThinkingConfig(body, budget)
}

// applyBoltHeaders sets Bolt-specific HTTP headers for Anthropic API.
func (e *BoltExecutor) applyBoltHeaders(r *http.Request, auth *cliproxyauth.Auth, apiKey string, stream bool, extraBetas []string) {
	r.Header.Del("Authorization")
	r.Header.Set("x-api-key", apiKey)
	r.Header.Set("Content-Type", "application/json")

	baseBetas := "claude-code-20250219,interleaved-thinking-2025-05-14,fine-grained-tool-streaming-2025-05-14"

	// Merge extra betas from request body
	if len(extraBetas) > 0 {
		existingSet := make(map[string]bool)
		for _, b := range strings.Split(baseBetas, ",") {
			existingSet[strings.TrimSpace(b)] = true
		}
		for _, beta := range extraBetas {
			beta = strings.TrimSpace(beta)
			if beta != "" && !existingSet[beta] {
				baseBetas += "," + beta
				existingSet[beta] = true
			}
		}
	}
	r.Header.Set("Anthropic-Beta", baseBetas)

	misc.EnsureHeader(r.Header, nil, "Anthropic-Version", "2023-06-01")
	misc.EnsureHeader(r.Header, nil, "User-Agent", "bolt.new/1.0")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

	if stream {
		r.Header.Set("Accept", "text/event-stream")
	} else {
		r.Header.Set("Accept", "application/json")
	}

	var attrs map[string]string
	if auth != nil {
		attrs = auth.Attributes
	}
	util.ApplyCustomHeadersFromAttrs(r, attrs)
}

// resolveUpstreamModel resolves model aliases to upstream model names.
func (e *BoltExecutor) resolveUpstreamModel(alias string, auth *cliproxyauth.Auth) string {
	trimmed := strings.TrimSpace(alias)
	if trimmed == "" {
		return ""
	}

	entry := e.resolveBoltConfig(auth)
	if entry == nil {
		return ""
	}

	normalizedModel, metadata := util.NormalizeThinkingModel(trimmed)

	candidates := []string{strings.TrimSpace(normalizedModel)}
	if !strings.EqualFold(normalizedModel, trimmed) {
		candidates = append(candidates, trimmed)
	}
	if original := util.ResolveOriginalModel(normalizedModel, metadata); original != "" && !strings.EqualFold(original, normalizedModel) {
		candidates = append(candidates, original)
	}

	for i := range entry.Models {
		model := entry.Models[i]
		name := strings.TrimSpace(model.Name)
		modelAlias := strings.TrimSpace(model.Alias)

		for _, candidate := range candidates {
			if candidate == "" {
				continue
			}
			if modelAlias != "" && strings.EqualFold(modelAlias, candidate) {
				if name != "" {
					return name
				}
				return candidate
			}
			if name != "" && strings.EqualFold(name, candidate) {
				return name
			}
		}
	}
	return ""
}

// resolveBoltConfig finds the matching configuration for the auth.
func (e *BoltExecutor) resolveBoltConfig(auth *cliproxyauth.Auth) *config.ClaudeKey {
	if auth == nil || e.cfg == nil {
		return nil
	}
	var attrKey, attrBase string
	if auth.Attributes != nil {
		attrKey = strings.TrimSpace(auth.Attributes["api_key"])
		attrBase = strings.TrimSpace(auth.Attributes["base_url"])
	}

	// Look for Bolt-specific config or fallback to Claude config
	for i := range e.cfg.ClaudeKey {
		entry := &e.cfg.ClaudeKey[i]
		cfgKey := strings.TrimSpace(entry.APIKey)
		cfgBase := strings.TrimSpace(entry.BaseURL)
		if attrKey != "" && attrBase != "" {
			if strings.EqualFold(cfgKey, attrKey) && strings.EqualFold(cfgBase, attrBase) {
				return entry
			}
			continue
		}
		if attrKey != "" && strings.EqualFold(cfgKey, attrKey) {
			if cfgBase == "" || strings.EqualFold(cfgBase, attrBase) {
				return entry
			}
		}
		if attrKey == "" && attrBase != "" && strings.EqualFold(cfgBase, attrBase) {
			return entry
		}
	}
	if attrKey != "" {
		for i := range e.cfg.ClaudeKey {
			entry := &e.cfg.ClaudeKey[i]
			if strings.EqualFold(strings.TrimSpace(entry.APIKey), attrKey) {
				return entry
			}
		}
	}
	return nil
}

// boltCreds extracts API key and base URL from auth.
func boltCreds(a *cliproxyauth.Auth) (apiKey, baseURL string) {
	if a == nil {
		return "", ""
	}
	if a.Attributes != nil {
		apiKey = a.Attributes["api_key"]
		baseURL = a.Attributes["base_url"]
	}
	if apiKey == "" && a.Metadata != nil {
		if v, ok := a.Metadata["api_key"].(string); ok {
			apiKey = v
		}
	}
	return
}

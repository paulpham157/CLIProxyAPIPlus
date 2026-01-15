package executor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	windsurfauth "github.com/router-for-me/CLIProxyAPI/v6/internal/auth/windsurf"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v6/sdk/translator"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/sjson"
)

const (
	windsurfBaseURL    = "https://proxy.codeium.com"
	windsurfChatPath   = "/v1/chat/completions"
	windsurfAuthType   = "windsurf"
	windsurfTokenTTL   = 25 * time.Minute
	tokenExpiryBuffer  = 5 * time.Minute
	maxScannerBufferSize = 20_971_520

	windsurfUserAgent = "Windsurf-CLI/1.0"
)

// WindsurfExecutor handles requests to the Windsurf AI API.
type WindsurfExecutor struct {
	cfg   *config.Config
	mu    sync.RWMutex
	cache map[string]*cachedWindsurfToken
}

// cachedWindsurfToken stores a cached Windsurf API token with its expiry.
type cachedWindsurfToken struct {
	token     string
	expiresAt time.Time
}

// NewWindsurfExecutor constructs a new executor instance.
func NewWindsurfExecutor(cfg *config.Config) *WindsurfExecutor {
	return &WindsurfExecutor{
		cfg:   cfg,
		cache: make(map[string]*cachedWindsurfToken),
	}
}

// Identifier implements ProviderExecutor.
func (e *WindsurfExecutor) Identifier() string { return windsurfAuthType }

// PrepareRequest implements ProviderExecutor.
func (e *WindsurfExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
	if req == nil {
		return nil
	}
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	apiToken, errToken := e.ensureAPIToken(ctx, auth)
	if errToken != nil {
		return errToken
	}
	e.applyHeaders(req, apiToken)
	return nil
}

// HttpRequest injects Windsurf credentials into the request and executes it.
func (e *WindsurfExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("windsurf executor: request is nil")
	}
	if ctx == nil {
		ctx = req.Context()
	}
	httpReq := req.WithContext(ctx)
	if errPrepare := e.PrepareRequest(httpReq, auth); errPrepare != nil {
		return nil, errPrepare
	}
	httpClient := newProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	return httpClient.Do(httpReq)
}

// Execute handles non-streaming requests to Windsurf.
func (e *WindsurfExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (resp cliproxyexecutor.Response, err error) {
	apiToken, errToken := e.ensureAPIToken(ctx, auth)
	if errToken != nil {
		return resp, errToken
	}

	reporter := newUsageReporter(ctx, e.Identifier(), req.Model, auth)
	defer reporter.trackFailure(ctx, &err)

	from := opts.SourceFormat
	to := sdktranslator.FromString("openai")
	originalPayload := bytes.Clone(req.Payload)
	if len(opts.OriginalRequest) > 0 {
		originalPayload = bytes.Clone(opts.OriginalRequest)
	}
	originalTranslated := sdktranslator.TranslateRequest(from, to, req.Model, originalPayload, false)
	body := sdktranslator.TranslateRequest(from, to, req.Model, bytes.Clone(req.Payload), false)
	body = e.normalizeModel(req.Model, body)
	body = applyPayloadConfigWithRoot(e.cfg, req.Model, to.String(), "", body, originalTranslated)
	body, _ = sjson.SetBytes(body, "stream", false)

	url := windsurfBaseURL + windsurfChatPath
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return resp, err
	}
	e.applyHeaders(httpReq, apiToken)

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
	defer func() {
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("windsurf executor: close response body error: %v", errClose)
		}
	}()

	recordAPIResponseMetadata(ctx, e.cfg, httpResp.StatusCode, httpResp.Header.Clone())

	if !isHTTPSuccess(httpResp.StatusCode) {
		data, _ := io.ReadAll(httpResp.Body)
		appendAPIResponseChunk(ctx, e.cfg, data)
		log.Debugf("windsurf executor: upstream error status: %d, body: %s", httpResp.StatusCode, summarizeErrorBody(httpResp.Header.Get("Content-Type"), data))
		err = statusErr{code: httpResp.StatusCode, msg: string(data)}
		return resp, err
	}

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		recordAPIResponseError(ctx, e.cfg, err)
		return resp, err
	}
	appendAPIResponseChunk(ctx, e.cfg, data)

	detail := parseOpenAIUsage(data)
	if detail.TotalTokens > 0 {
		reporter.publish(ctx, detail)
	}

	var param any
	converted := sdktranslator.TranslateNonStream(ctx, to, from, req.Model, bytes.Clone(opts.OriginalRequest), body, data, &param)
	resp = cliproxyexecutor.Response{Payload: []byte(converted)}
	reporter.ensurePublished(ctx)
	return resp, nil
}

// ExecuteStream handles streaming requests to Windsurf.
func (e *WindsurfExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (stream <-chan cliproxyexecutor.StreamChunk, err error) {
	apiToken, errToken := e.ensureAPIToken(ctx, auth)
	if errToken != nil {
		return nil, errToken
	}

	reporter := newUsageReporter(ctx, e.Identifier(), req.Model, auth)
	defer reporter.trackFailure(ctx, &err)

	from := opts.SourceFormat
	to := sdktranslator.FromString("openai")
	originalPayload := bytes.Clone(req.Payload)
	if len(opts.OriginalRequest) > 0 {
		originalPayload = bytes.Clone(opts.OriginalRequest)
	}
	originalTranslated := sdktranslator.TranslateRequest(from, to, req.Model, originalPayload, false)
	body := sdktranslator.TranslateRequest(from, to, req.Model, bytes.Clone(req.Payload), true)
	body = e.normalizeModel(req.Model, body)
	body = applyPayloadConfigWithRoot(e.cfg, req.Model, to.String(), "", body, originalTranslated)
	body, _ = sjson.SetBytes(body, "stream", true)
	body, _ = sjson.SetBytes(body, "stream_options.include_usage", true)

	url := windsurfBaseURL + windsurfChatPath
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	e.applyHeaders(httpReq, apiToken)

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

	if !isHTTPSuccess(httpResp.StatusCode) {
		data, readErr := io.ReadAll(httpResp.Body)
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("windsurf executor: close response body error: %v", errClose)
		}
		if readErr != nil {
			recordAPIResponseError(ctx, e.cfg, readErr)
			return nil, readErr
		}
		appendAPIResponseChunk(ctx, e.cfg, data)
		log.Debugf("windsurf executor: upstream error status: %d, body: %s", httpResp.StatusCode, summarizeErrorBody(httpResp.Header.Get("Content-Type"), data))
		err = statusErr{code: httpResp.StatusCode, msg: string(data)}
		return nil, err
	}

	out := make(chan cliproxyexecutor.StreamChunk)
	stream = out

	go func() {
		defer close(out)
		defer func() {
			if errClose := httpResp.Body.Close(); errClose != nil {
				log.Errorf("windsurf executor: close response body error: %v", errClose)
			}
		}()

		scanner := bufio.NewScanner(httpResp.Body)
		scanner.Buffer(nil, maxScannerBufferSize)
		var param any

		for scanner.Scan() {
			line := scanner.Bytes()
			appendAPIResponseChunk(ctx, e.cfg, line)

			if bytes.HasPrefix(line, dataTag) {
				data := bytes.TrimSpace(line[5:])
				if bytes.Equal(data, []byte("[DONE]")) {
					continue
				}
				if detail, ok := parseOpenAIStreamUsage(line); ok {
					reporter.publish(ctx, detail)
				}
			}

			chunks := sdktranslator.TranslateStream(ctx, to, from, req.Model, bytes.Clone(opts.OriginalRequest), body, bytes.Clone(line), &param)
			for i := range chunks {
				out <- cliproxyexecutor.StreamChunk{Payload: []byte(chunks[i])}
			}
		}

		if errScan := scanner.Err(); errScan != nil {
			recordAPIResponseError(ctx, e.cfg, errScan)
			reporter.publishFailure(ctx)
			out <- cliproxyexecutor.StreamChunk{Err: errScan}
		} else {
			reporter.ensurePublished(ctx)
		}
	}()

	return stream, nil
}

// CountTokens is not supported for Windsurf.
func (e *WindsurfExecutor) CountTokens(_ context.Context, _ *cliproxyauth.Auth, _ cliproxyexecutor.Request, _ cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	return cliproxyexecutor.Response{}, statusErr{code: http.StatusNotImplemented, msg: "count tokens not supported for windsurf"}
}

// Refresh validates the Windsurf token is still working.
func (e *WindsurfExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	if auth == nil {
		return nil, statusErr{code: http.StatusUnauthorized, msg: "missing auth"}
	}

	accessToken := metaStringValue(auth.Metadata, "access_token")
	if accessToken == "" {
		return auth, nil
	}

	windsurfAuth := windsurfauth.NewWindsurfAuth(e.cfg)
	valid, _, err := windsurfAuth.ValidateToken(ctx, accessToken)
	if err != nil || !valid {
		return nil, statusErr{code: http.StatusUnauthorized, msg: fmt.Sprintf("windsurf token validation failed: %v", err)}
	}

	return auth, nil
}

// ensureAPIToken gets or refreshes the Windsurf API token.
func (e *WindsurfExecutor) ensureAPIToken(ctx context.Context, auth *cliproxyauth.Auth) (string, error) {
	if auth == nil {
		return "", statusErr{code: http.StatusUnauthorized, msg: "missing auth"}
	}

	accessToken := metaStringValue(auth.Metadata, "access_token")
	if accessToken == "" {
		return "", statusErr{code: http.StatusUnauthorized, msg: "missing windsurf access token"}
	}

	e.mu.RLock()
	if cached, ok := e.cache[accessToken]; ok && cached.expiresAt.After(time.Now().Add(tokenExpiryBuffer)) {
		e.mu.RUnlock()
		return cached.token, nil
	}
	e.mu.RUnlock()

	expiresAt := time.Now().Add(windsurfTokenTTL)
	e.mu.Lock()
	e.cache[accessToken] = &cachedWindsurfToken{
		token:     accessToken,
		expiresAt: expiresAt,
	}
	e.mu.Unlock()

	return accessToken, nil
}

// applyHeaders sets the required headers for Windsurf API requests.
func (e *WindsurfExecutor) applyHeaders(r *http.Request, apiToken string) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+apiToken)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("User-Agent", windsurfUserAgent)
	r.Header.Set("X-Request-Id", uuid.NewString())
}

// normalizeModel is a no-op as Windsurf accepts model names directly.
func (e *WindsurfExecutor) normalizeModel(_ string, body []byte) []byte {
	return body
}

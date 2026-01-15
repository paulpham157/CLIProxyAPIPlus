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
	continueauth "github.com/router-for-me/CLIProxyAPI/v6/internal/auth/continue"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v6/sdk/translator"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/sjson"
)

const (
	continueBaseURL       = "https://api.continue.dev"
	continueChatPath      = "/v1/chat/completions"
	continueAuthType      = "continue"
	continueTokenCacheTTL = 25 * time.Minute
	tokenExpiryBuffer     = 5 * time.Minute
	maxScannerBufferSize  = 20_971_520
)

type ContinueExecutor struct {
	cfg   *config.Config
	mu    sync.RWMutex
	cache map[string]*cachedAPIToken
}

func NewContinueExecutor(cfg *config.Config) *ContinueExecutor {
	return &ContinueExecutor{
		cfg:   cfg,
		cache: make(map[string]*cachedAPIToken),
	}
}

func (e *ContinueExecutor) Identifier() string { return continueAuthType }

func (e *ContinueExecutor) PrepareRequest(req *http.Request, auth *cliproxyauth.Auth) error {
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

func (e *ContinueExecutor) HttpRequest(ctx context.Context, auth *cliproxyauth.Auth, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("continue executor: request is nil")
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

func (e *ContinueExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (resp cliproxyexecutor.Response, err error) {
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

	url := continueBaseURL + continueChatPath
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
			log.Errorf("continue executor: close response body error: %v", errClose)
		}
	}()

	recordAPIResponseMetadata(ctx, e.cfg, httpResp.StatusCode, httpResp.Header.Clone())

	if !isHTTPSuccess(httpResp.StatusCode) {
		data, _ := io.ReadAll(httpResp.Body)
		appendAPIResponseChunk(ctx, e.cfg, data)
		log.Debugf("continue executor: upstream error status: %d, body: %s", httpResp.StatusCode, summarizeErrorBody(httpResp.Header.Get("Content-Type"), data))
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

func (e *ContinueExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (stream <-chan cliproxyexecutor.StreamChunk, err error) {
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
	originalTranslated := sdktranslator.TranslateRequest(from, to, req.Model, originalPayload, true)
	body := sdktranslator.TranslateRequest(from, to, req.Model, bytes.Clone(req.Payload), true)
	body = e.normalizeModel(req.Model, body)
	body = applyPayloadConfigWithRoot(e.cfg, req.Model, to.String(), "", body, originalTranslated)
	body, _ = sjson.SetBytes(body, "stream", true)

	url := continueBaseURL + continueChatPath
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
		data, _ := io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()
		appendAPIResponseChunk(ctx, e.cfg, data)
		log.Debugf("continue executor: upstream error status: %d, body: %s", httpResp.StatusCode, summarizeErrorBody(httpResp.Header.Get("Content-Type"), data))
		err = statusErr{code: httpResp.StatusCode, msg: string(data)}
		return nil, err
	}

	outCh := make(chan cliproxyexecutor.StreamChunk, 100)
	go e.streamResponse(ctx, auth, httpResp, outCh, from, to, req.Model, opts.OriginalRequest, body, reporter)
	return outCh, nil
}

func (e *ContinueExecutor) streamResponse(ctx context.Context, auth *cliproxyauth.Auth, httpResp *http.Response, outCh chan cliproxyexecutor.StreamChunk, from, to sdktranslator.Format, model string, originalRequest, body []byte, reporter *usageReporter) {
	defer func() {
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("continue executor: close stream response body error: %v", errClose)
		}
		close(outCh)
	}()

	requestID := uuid.New().String()
	scanner := bufio.NewScanner(httpResp.Body)
	buf := make([]byte, 0, bufio.MaxScanTokenSize)
	scanner.Buffer(buf, maxScannerBufferSize)

	var param any
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			reporter.markCancelled(ctx)
			reporter.ensurePublished(ctx)
			return
		default:
		}

		line := scanner.Bytes()
		appendAPIResponseChunk(ctx, e.cfg, line)

		parts := sdktranslator.TranslateStream(ctx, to, from, model, bytes.Clone(originalRequest), body, line, &param)
		for i := range parts {
			select {
			case outCh <- cliproxyexecutor.StreamChunk{
				Data:      []byte(parts[i]),
				RequestID: requestID,
			}:
			case <-ctx.Done():
				reporter.markCancelled(ctx)
				reporter.ensurePublished(ctx)
				return
			}
		}
	}

	if errScan := scanner.Err(); errScan != nil {
		recordAPIResponseError(ctx, e.cfg, errScan)
		log.Errorf("continue executor: scanner error: %v", errScan)
	}
	reporter.ensurePublished(ctx)
}

func (e *ContinueExecutor) ensureAPIToken(ctx context.Context, auth *cliproxyauth.Auth) (string, error) {
	if auth == nil {
		return "", fmt.Errorf("continue executor: auth is nil")
	}

	e.mu.RLock()
	cached, found := e.cache[auth.ID]
	e.mu.RUnlock()

	now := time.Now()
	if found && cached != nil && now.Before(cached.expiresAt.Add(-tokenExpiryBuffer)) {
		return cached.token, nil
	}

	storage := &continueauth.ContinueTokenStorage{}
	if err := auth.LoadStorage(storage); err != nil {
		return "", fmt.Errorf("continue executor: failed to load token storage: %w", err)
	}

	authSvc := continueauth.NewContinueAuth(e.cfg)
	apiToken, err := authSvc.GetContinueAPIToken(ctx, storage.AccessToken)
	if err != nil {
		return "", fmt.Errorf("continue executor: failed to get API token: %w", err)
	}

	if apiToken.Token == "" {
		return "", fmt.Errorf("continue executor: empty API token")
	}

	expiry := now.Add(continueTokenCacheTTL)
	if apiToken.ExpiresAt > 0 {
		expiry = time.Unix(apiToken.ExpiresAt, 0)
	}

	e.mu.Lock()
	e.cache[auth.ID] = &cachedAPIToken{
		token:     apiToken.Token,
		expiresAt: expiry,
	}
	e.mu.Unlock()

	return apiToken.Token, nil
}

func (e *ContinueExecutor) applyHeaders(req *http.Request, apiToken string) {
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

func (e *ContinueExecutor) normalizeModel(model string, body []byte) []byte {
	body, _ = sjson.SetBytes(body, "model", model)
	return body
}

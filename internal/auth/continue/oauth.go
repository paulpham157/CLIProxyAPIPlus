package continueauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/util"
	log "github.com/sirupsen/logrus"
)

const (
	continueClientID      = "continue-dev-client"
	continueDeviceCodeURL = "https://auth.continue.dev/device/code"
	continueTokenURL      = "https://auth.continue.dev/oauth/token"
	continueUserInfoURL   = "https://api.continue.dev/user"
	defaultPollInterval   = 5 * time.Second
	maxPollDuration       = 15 * time.Minute
)

type DeviceFlowClient struct {
	httpClient *http.Client
	cfg        *config.Config
}

func NewDeviceFlowClient(cfg *config.Config) *DeviceFlowClient {
	client := &http.Client{Timeout: 30 * time.Second}
	if cfg != nil {
		client = util.SetProxy(&cfg.SDKConfig, client)
	}
	return &DeviceFlowClient{
		httpClient: client,
		cfg:        cfg,
	}
}

func (c *DeviceFlowClient) RequestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", continueClientID)
	data.Set("scope", "user:email")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, continueDeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, NewAuthenticationError(ErrDeviceCodeFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAuthenticationError(ErrDeviceCodeFailed, err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("continue device code: close body error: %v", errClose)
		}
	}()

	if !isHTTPSuccess(resp.StatusCode) {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, NewAuthenticationError(ErrDeviceCodeFailed, fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes)))
	}

	var deviceCode DeviceCodeResponse
	if err = json.NewDecoder(resp.Body).Decode(&deviceCode); err != nil {
		return nil, NewAuthenticationError(ErrDeviceCodeFailed, err)
	}

	return &deviceCode, nil
}

func (c *DeviceFlowClient) PollForToken(ctx context.Context, deviceCode *DeviceCodeResponse) (*ContinueTokenData, error) {
	if deviceCode == nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, fmt.Errorf("device code is nil"))
	}

	interval := time.Duration(deviceCode.Interval) * time.Second
	if interval < defaultPollInterval {
		interval = defaultPollInterval
	}

	deadline := time.Now().Add(maxPollDuration)
	if deviceCode.ExpiresIn > 0 {
		codeDeadline := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)
		if codeDeadline.Before(deadline) {
			deadline = codeDeadline
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, NewAuthenticationError(ErrPollingTimeout, ctx.Err())
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, ErrPollingTimeout
			}

			token, err := c.exchangeDeviceCode(ctx, deviceCode.DeviceCode)
			if err != nil {
				var authErr *AuthenticationError
				if errors.As(err, &authErr) {
					switch authErr.Type {
					case ErrAuthorizationPending.Type:
						continue
					case ErrSlowDown.Type:
						interval += 5 * time.Second
						ticker.Reset(interval)
						continue
					case ErrDeviceCodeExpired.Type:
						return nil, err
					case ErrAccessDenied.Type:
						return nil, err
					}
				}
				return nil, err
			}
			return token, nil
		}
	}
}

func (c *DeviceFlowClient) exchangeDeviceCode(ctx context.Context, deviceCode string) (*ContinueTokenData, error) {
	data := url.Values{}
	data.Set("client_id", continueClientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, continueTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("continue token exchange: close body error: %v", errClose)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}

	var oauthResp struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		Scope            string `json:"scope"`
	}

	if err = json.Unmarshal(bodyBytes, &oauthResp); err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}

	if oauthResp.Error != "" {
		switch oauthResp.Error {
		case "authorization_pending":
			return nil, ErrAuthorizationPending
		case "slow_down":
			return nil, ErrSlowDown
		case "expired_token":
			return nil, ErrDeviceCodeExpired
		case "access_denied":
			return nil, ErrAccessDenied
		default:
			return nil, NewOAuthError(oauthResp.Error, oauthResp.ErrorDescription, resp.StatusCode)
		}
	}

	if oauthResp.AccessToken == "" {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, fmt.Errorf("empty access token"))
	}

	return &ContinueTokenData{
		AccessToken: oauthResp.AccessToken,
		TokenType:   oauthResp.TokenType,
		Scope:       oauthResp.Scope,
	}, nil
}

func (c *DeviceFlowClient) FetchUserInfo(ctx context.Context, accessToken string) (string, error) {
	if accessToken == "" {
		return "", NewAuthenticationError(ErrUserInfoFailed, fmt.Errorf("access token is empty"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, continueUserInfoURL, nil)
	if err != nil {
		return "", NewAuthenticationError(ErrUserInfoFailed, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", NewAuthenticationError(ErrUserInfoFailed, err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("continue user info: close body error: %v", errClose)
		}
	}()

	if !isHTTPSuccess(resp.StatusCode) {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", NewAuthenticationError(ErrUserInfoFailed, fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes)))
	}

	var userInfo struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", NewAuthenticationError(ErrUserInfoFailed, err)
	}

	if userInfo.Username == "" && userInfo.Email == "" {
		return "", NewAuthenticationError(ErrUserInfoFailed, fmt.Errorf("empty username and email"))
	}

	if userInfo.Username != "" {
		return userInfo.Username, nil
	}
	return userInfo.Email, nil
}

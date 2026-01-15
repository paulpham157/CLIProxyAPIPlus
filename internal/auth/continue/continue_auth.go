package continueauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/util"
	log "github.com/sirupsen/logrus"
)

const (
	continueAPIEndpoint = "https://api.continue.dev"
)

type ContinueAPIToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type ContinueAuth struct {
	httpClient   *http.Client
	deviceClient *DeviceFlowClient
	cfg          *config.Config
}

func NewContinueAuth(cfg *config.Config) *ContinueAuth {
	return &ContinueAuth{
		httpClient:   util.SetProxy(&cfg.SDKConfig, &http.Client{Timeout: 30 * time.Second}),
		deviceClient: NewDeviceFlowClient(cfg),
		cfg:          cfg,
	}
}

func (c *ContinueAuth) StartDeviceFlow(ctx context.Context) (*DeviceCodeResponse, error) {
	return c.deviceClient.RequestDeviceCode(ctx)
}

func (c *ContinueAuth) WaitForAuthorization(ctx context.Context, deviceCode *DeviceCodeResponse) (*ContinueAuthBundle, error) {
	tokenData, err := c.deviceClient.PollForToken(ctx, deviceCode)
	if err != nil {
		return nil, err
	}

	username, err := c.deviceClient.FetchUserInfo(ctx, tokenData.AccessToken)
	if err != nil {
		log.Warnf("continue: failed to fetch user info: %v", err)
		username = "unknown"
	}

	return &ContinueAuthBundle{
		TokenData: tokenData,
		Username:  username,
	}, nil
}

func (c *ContinueAuth) GetContinueAPIToken(ctx context.Context, accessToken string) (*ContinueAPIToken, error) {
	if accessToken == "" {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, fmt.Errorf("access token is empty"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, continueAPIEndpoint+"/auth/token", nil)
	if err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("continue api token: close body error: %v", errClose)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}

	if !isHTTPSuccess(resp.StatusCode) {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed,
			fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes)))
	}

	var apiToken ContinueAPIToken
	if err = json.Unmarshal(bodyBytes, &apiToken); err != nil {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, err)
	}

	if apiToken.Token == "" {
		return nil, NewAuthenticationError(ErrTokenExchangeFailed, fmt.Errorf("empty continue api token"))
	}

	return &apiToken, nil
}

func (c *ContinueAuth) ValidateToken(ctx context.Context, accessToken string) (bool, string, error) {
	if accessToken == "" {
		return false, "", nil
	}

	username, err := c.deviceClient.FetchUserInfo(ctx, accessToken)
	if err != nil {
		return false, "", err
	}

	return true, username, nil
}

func (c *ContinueAuth) CreateTokenStorage(bundle *ContinueAuthBundle) *ContinueTokenStorage {
	return &ContinueTokenStorage{
		AccessToken: bundle.TokenData.AccessToken,
		TokenType:   bundle.TokenData.TokenType,
		Scope:       bundle.TokenData.Scope,
		Username:    bundle.Username,
		Type:        "continue",
	}
}

func (c *ContinueAuth) LoadAndValidateToken(ctx context.Context, storage *ContinueTokenStorage) (bool, error) {
	if storage == nil || storage.AccessToken == "" {
		return false, fmt.Errorf("no token available")
	}

	apiToken, err := c.GetContinueAPIToken(ctx, storage.AccessToken)
	if err != nil {
		return false, err
	}

	if apiToken.ExpiresAt > 0 && time.Now().Unix() >= apiToken.ExpiresAt {
		return false, fmt.Errorf("continue api token expired")
	}

	return true, nil
}

func (c *ContinueAuth) GetAPIEndpoint() string {
	return continueAPIEndpoint
}

func isHTTPSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

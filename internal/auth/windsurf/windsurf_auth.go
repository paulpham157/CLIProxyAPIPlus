// Package windsurf provides authentication and token management for Windsurf AI API.
// It handles the OAuth2 authentication flow for secure access to Windsurf's AI backend.
package windsurf

import (
	"bytes"
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
	// windsurfAPIEndpoint is the base URL for making API requests to Windsurf.
	windsurfAPIEndpoint = "https://proxy.codeium.com"
	// windsurfChatPath is the chat completions endpoint path.
	windsurfChatPath = "/v1/chat/completions"

	// Common HTTP header values for Windsurf API requests.
	windsurfUserAgent = "Windsurf-CLI/1.0"
)

// WindsurfAPIToken represents the Windsurf API token response.
type WindsurfAPIToken struct {
	// Token is the JWT token for authenticating with the Windsurf API.
	Token string `json:"token"`
	// ExpiresAt is the Unix timestamp when the token expires.
	ExpiresAt int64 `json:"expires_at"`
}

// WindsurfAuth handles Windsurf authentication flow.
// It provides methods for OAuth authentication and token management.
type WindsurfAuth struct {
	httpClient   *http.Client
	deviceClient *DeviceFlowClient
	cfg          *config.Config
}

// NewWindsurfAuth creates a new WindsurfAuth service instance.
// It initializes an HTTP client with proxy settings from the provided configuration.
func NewWindsurfAuth(cfg *config.Config) *WindsurfAuth {
	return &WindsurfAuth{
		httpClient:   util.SetProxy(&cfg.SDKConfig, &http.Client{Timeout: 30 * time.Second}),
		deviceClient: NewDeviceFlowClient(cfg),
		cfg:          cfg,
	}
}

// StartDeviceFlow initiates the device flow authentication.
// Returns the device code response containing the user code and verification URI.
func (c *WindsurfAuth) StartDeviceFlow(ctx context.Context) (*DeviceCodeResponse, error) {
	return c.deviceClient.RequestDeviceCode(ctx)
}

// WaitForAuthorization polls for user authorization and returns the auth bundle.
func (c *WindsurfAuth) WaitForAuthorization(ctx context.Context, deviceCode *DeviceCodeResponse) (*WindsurfAuthBundle, error) {
	tokenData, err := c.deviceClient.PollForToken(ctx, deviceCode)
	if err != nil {
		return nil, err
	}

	// Fetch the user info
	userInfo, err := c.deviceClient.FetchUserInfo(ctx, tokenData.AccessToken)
	if err != nil {
		log.Warnf("windsurf: failed to fetch user info: %v", err)
		userInfo = "unknown"
	}

	return &WindsurfAuthBundle{
		TokenData: *tokenData,
		UserInfo:  userInfo,
	}, nil
}

// ValidateToken checks if an access token is valid by attempting to fetch user info.
func (c *WindsurfAuth) ValidateToken(ctx context.Context, accessToken string) (bool, string, error) {
	if accessToken == "" {
		return false, "", nil
	}

	userInfo, err := c.deviceClient.FetchUserInfo(ctx, accessToken)
	if err != nil {
		return false, "", err
	}

	return true, userInfo, nil
}

// CreateTokenStorage creates a new WindsurfTokenStorage from auth bundle.
func (c *WindsurfAuth) CreateTokenStorage(bundle *WindsurfAuthBundle) *WindsurfTokenStorage {
	return &WindsurfTokenStorage{
		AccessToken:  bundle.TokenData.AccessToken,
		RefreshToken: bundle.TokenData.RefreshToken,
		TokenType:    bundle.TokenData.TokenType,
		UserInfo:     bundle.UserInfo,
		Type:         "windsurf",
	}
}

// LoadAndValidateToken loads a token from storage and validates it.
// Returns the storage if valid, or an error if the token is invalid or expired.
func (c *WindsurfAuth) LoadAndValidateToken(ctx context.Context, storage *WindsurfTokenStorage) (bool, error) {
	if storage == nil || storage.AccessToken == "" {
		return false, fmt.Errorf("no token available")
	}

	// Validate the token
	valid, _, err := c.ValidateToken(ctx, storage.AccessToken)
	if err != nil {
		return false, err
	}

	return valid, nil
}

// GetAPIEndpoint returns the Windsurf API endpoint URL.
func (c *WindsurfAuth) GetAPIEndpoint() string {
	return windsurfAPIEndpoint
}

// MakeAuthenticatedRequest creates an authenticated HTTP request to the Windsurf API.
func (c *WindsurfAuth) MakeAuthenticatedRequest(ctx context.Context, method, url string, body io.Reader, accessToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", windsurfUserAgent)

	return req, nil
}

// RefreshToken refreshes the access token using the refresh token if available.
func (c *WindsurfAuth) RefreshToken(ctx context.Context, refreshToken string) (*WindsurfTokenData, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	reqBody := map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     windsurfClientID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", windsurfTokenURL, io.NopCloser(bytes.NewReader(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &WindsurfTokenData{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
	}, nil
}

// buildChatCompletionURL builds the URL for chat completions API.
func buildChatCompletionURL() string {
	return windsurfAPIEndpoint + windsurfChatPath
}

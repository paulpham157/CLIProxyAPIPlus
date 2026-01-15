// Package cursor provides authentication and token management for Cursor AI API.
// It handles the OAuth2 authentication flow for secure access to Cursor's AI backend.
package cursor

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
	// cursorAPIEndpoint is the base URL for making API requests to Cursor.
	cursorAPIEndpoint = "https://api.cursor.sh"
	// cursorChatPath is the chat completions endpoint path.
	cursorChatPath = "/v1/chat/completions"

	// Common HTTP header values for Cursor API requests.
	cursorUserAgent = "Cursor-CLI/1.0"
)

// CursorAPIToken represents the Cursor API token response.
type CursorAPIToken struct {
	// Token is the JWT token for authenticating with the Cursor API.
	Token string `json:"token"`
	// ExpiresAt is the Unix timestamp when the token expires.
	ExpiresAt int64 `json:"expires_at"`
}

// CursorAuth handles Cursor authentication flow.
// It provides methods for OAuth authentication and token management.
type CursorAuth struct {
	httpClient   *http.Client
	deviceClient *DeviceFlowClient
	cfg          *config.Config
}

// NewCursorAuth creates a new CursorAuth service instance.
// It initializes an HTTP client with proxy settings from the provided configuration.
func NewCursorAuth(cfg *config.Config) *CursorAuth {
	return &CursorAuth{
		httpClient:   util.SetProxy(&cfg.SDKConfig, &http.Client{Timeout: 30 * time.Second}),
		deviceClient: NewDeviceFlowClient(cfg),
		cfg:          cfg,
	}
}

// StartDeviceFlow initiates the device flow authentication.
// Returns the device code response containing the user code and verification URI.
func (c *CursorAuth) StartDeviceFlow(ctx context.Context) (*DeviceCodeResponse, error) {
	return c.deviceClient.RequestDeviceCode(ctx)
}

// WaitForAuthorization polls for user authorization and returns the auth bundle.
func (c *CursorAuth) WaitForAuthorization(ctx context.Context, deviceCode *DeviceCodeResponse) (*CursorAuthBundle, error) {
	tokenData, err := c.deviceClient.PollForToken(ctx, deviceCode)
	if err != nil {
		return nil, err
	}

	// Fetch the user info
	userInfo, err := c.deviceClient.FetchUserInfo(ctx, tokenData.AccessToken)
	if err != nil {
		log.Warnf("cursor: failed to fetch user info: %v", err)
		userInfo = "unknown"
	}

	return &CursorAuthBundle{
		TokenData: tokenData,
		UserInfo:  userInfo,
	}, nil
}

// ValidateToken checks if an access token is valid by attempting to fetch user info.
func (c *CursorAuth) ValidateToken(ctx context.Context, accessToken string) (bool, string, error) {
	if accessToken == "" {
		return false, "", nil
	}

	userInfo, err := c.deviceClient.FetchUserInfo(ctx, accessToken)
	if err != nil {
		return false, "", err
	}

	return true, userInfo, nil
}

// CreateTokenStorage creates a new CursorTokenStorage from auth bundle.
func (c *CursorAuth) CreateTokenStorage(bundle *CursorAuthBundle) *CursorTokenStorage {
	return &CursorTokenStorage{
		AccessToken: bundle.TokenData.AccessToken,
		TokenType:   bundle.TokenData.TokenType,
		UserInfo:    bundle.UserInfo,
		Type:        "cursor",
	}
}

// LoadAndValidateToken loads a token from storage and validates it.
// Returns the storage if valid, or an error if the token is invalid or expired.
func (c *CursorAuth) LoadAndValidateToken(ctx context.Context, storage *CursorTokenStorage) (bool, error) {
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

// GetAPIEndpoint returns the Cursor API endpoint URL.
func (c *CursorAuth) GetAPIEndpoint() string {
	return cursorAPIEndpoint
}

// MakeAuthenticatedRequest creates an authenticated HTTP request to the Cursor API.
func (c *CursorAuth) MakeAuthenticatedRequest(ctx context.Context, method, url string, body io.Reader, accessToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", cursorUserAgent)

	return req, nil
}

// RefreshToken refreshes the access token using the refresh token if available.
func (c *CursorAuth) RefreshToken(ctx context.Context, refreshToken string) (*CursorTokenData, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	reqBody := map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cursorAPIEndpoint+"/oauth/token", io.NopCloser(bytes.NewReader(jsonBody)))
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

	return &CursorTokenData{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
	}, nil
}

// buildChatCompletionURL builds the URL for chat completions API.
func buildChatCompletionURL() string {
	return cursorAPIEndpoint + cursorChatPath
}

// isHTTPSuccess checks if the status code indicates success (2xx).
func isHTTPSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

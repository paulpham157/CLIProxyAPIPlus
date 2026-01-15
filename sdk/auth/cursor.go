package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/auth/cursor"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

// CursorAuthenticator implements the OAuth device flow login for Cursor AI.
type CursorAuthenticator struct{}

// NewCursorAuthenticator constructs a new Cursor authenticator.
func NewCursorAuthenticator() Authenticator {
	return &CursorAuthenticator{}
}

// Provider returns the provider key for cursor.
func (CursorAuthenticator) Provider() string {
	return "cursor"
}

// RefreshLead returns nil since Cursor OAuth tokens don't expire in the traditional sense.
func (CursorAuthenticator) RefreshLead() *time.Duration {
	return nil
}

// Login initiates the Cursor device flow authentication.
func (a CursorAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cliproxy auth: configuration is required")
	}
	if opts == nil {
		opts = &LoginOptions{}
	}

	authSvc := cursor.NewCursorAuth(cfg)

	fmt.Println("Starting Cursor AI authentication...")
	deviceCode, err := authSvc.StartDeviceFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("cursor: failed to start device flow: %w", err)
	}

	fmt.Printf("\nTo authenticate, please visit: %s\n", deviceCode.VerificationURI)
	fmt.Printf("And enter the code: %s\n\n", deviceCode.UserCode)

	if !opts.NoBrowser {
		if browser.IsAvailable() {
			if errOpen := browser.OpenURL(deviceCode.VerificationURI); errOpen != nil {
				log.Warnf("Failed to open browser automatically: %v", errOpen)
			}
		}
	}

	fmt.Println("Waiting for Cursor authorization...")
	fmt.Printf("(This will timeout in %d seconds if not authorized)\n", deviceCode.ExpiresIn)

	authBundle, err := authSvc.WaitForAuthorization(ctx, deviceCode)
	if err != nil {
		errMsg := cursor.GetUserFriendlyMessage(err)
		return nil, fmt.Errorf("cursor: %s", errMsg)
	}

	fmt.Println("Verifying Cursor access...")

	tokenStorage := authSvc.CreateTokenStorage(authBundle)

	metadata := map[string]any{
		"type":         "cursor",
		"user_info":    authBundle.UserInfo,
		"access_token": authBundle.TokenData.AccessToken,
		"token_type":   authBundle.TokenData.TokenType,
		"timestamp":    time.Now().UnixMilli(),
	}

	if authBundle.TokenData.RefreshToken != "" {
		metadata["refresh_token"] = authBundle.TokenData.RefreshToken
	}

	fileName := fmt.Sprintf("cursor-%s.json", authBundle.UserInfo)

	fmt.Printf("\nCursor AI authentication successful for user: %s\n", authBundle.UserInfo)

	return &coreauth.Auth{
		ID:       fileName,
		Provider: a.Provider(),
		FileName: fileName,
		Label:    authBundle.UserInfo,
		Storage:  tokenStorage,
		Metadata: metadata,
	}, nil
}

// RefreshCursorToken validates and returns the current token status.
func RefreshCursorToken(ctx context.Context, cfg *config.Config, storage *cursor.CursorTokenStorage) error {
	if storage == nil || storage.AccessToken == "" {
		return fmt.Errorf("no token available")
	}

	authSvc := cursor.NewCursorAuth(cfg)

	valid, _, err := authSvc.ValidateToken(ctx, storage.AccessToken)
	if err != nil || !valid {
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}

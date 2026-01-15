package auth

import (
	"context"
	"fmt"
	"time"

	continueauth "github.com/router-for-me/CLIProxyAPI/v6/internal/auth/continue"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

type ContinueAuthenticator struct{}

func NewContinueAuthenticator() Authenticator {
	return &ContinueAuthenticator{}
}

func (ContinueAuthenticator) Provider() string {
	return "continue"
}

func (ContinueAuthenticator) RefreshLead() *time.Duration {
	return nil
}

func (a ContinueAuthenticator) Login(ctx context.Context, cfg *config.Config, opts *LoginOptions) (*coreauth.Auth, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cliproxy auth: configuration is required")
	}
	if opts == nil {
		opts = &LoginOptions{}
	}

	authSvc := continueauth.NewContinueAuth(cfg)

	fmt.Println("Starting Continue.dev authentication...")
	deviceCode, err := authSvc.StartDeviceFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("continue: failed to start device flow: %w", err)
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

	fmt.Println("Waiting for Continue.dev authorization...")
	fmt.Printf("(This will timeout in %d seconds if not authorized)\n", deviceCode.ExpiresIn)

	authBundle, err := authSvc.WaitForAuthorization(ctx, deviceCode)
	if err != nil {
		errMsg := continueauth.GetUserFriendlyMessage(err)
		return nil, fmt.Errorf("continue: %s", errMsg)
	}

	fmt.Println("Verifying Continue.dev access...")
	apiToken, err := authSvc.GetContinueAPIToken(ctx, authBundle.TokenData.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("continue: failed to verify Continue.dev access: %w", err)
	}

	tokenStorage := authSvc.CreateTokenStorage(authBundle)

	metadata := map[string]any{
		"type":         "continue",
		"username":     authBundle.Username,
		"access_token": authBundle.TokenData.AccessToken,
		"token_type":   authBundle.TokenData.TokenType,
		"scope":        authBundle.TokenData.Scope,
		"timestamp":    time.Now().UnixMilli(),
	}

	if apiToken.ExpiresAt > 0 {
		metadata["api_token_expires_at"] = apiToken.ExpiresAt
	}

	fileName := fmt.Sprintf("continue-%s.json", authBundle.Username)

	fmt.Printf("\nContinue.dev authentication successful for user: %s\n", authBundle.Username)

	return &coreauth.Auth{
		ID:       fileName,
		Provider: a.Provider(),
		FileName: fileName,
		Label:    authBundle.Username,
		Storage:  tokenStorage,
		Metadata: metadata,
	}, nil
}

func RefreshContinueToken(ctx context.Context, cfg *config.Config, storage *continueauth.ContinueTokenStorage) error {
	if storage == nil || storage.AccessToken == "" {
		return fmt.Errorf("no token available")
	}

	authSvc := continueauth.NewContinueAuth(cfg)

	_, err := authSvc.GetContinueAPIToken(ctx, storage.AccessToken)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}

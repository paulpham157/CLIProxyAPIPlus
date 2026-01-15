package continueauth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/misc"
)

type ContinueTokenStorage struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	Username    string `json:"username"`
	Type        string `json:"type"`
}

type ContinueTokenData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type ContinueAuthBundle struct {
	TokenData *ContinueTokenData
	Username  string
}

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (ts *ContinueTokenStorage) SaveTokenToFile(authFilePath string) error {
	misc.LogSavingCredentials(authFilePath)
	ts.Type = "continue"
	if err := os.MkdirAll(filepath.Dir(authFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	f, err := os.Create(authFilePath)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err = json.NewEncoder(f).Encode(ts); err != nil {
		return fmt.Errorf("failed to write token to file: %w", err)
	}
	return nil
}

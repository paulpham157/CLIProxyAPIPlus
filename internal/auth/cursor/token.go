package cursor

// CursorTokenData holds the OAuth token response from Cursor.
type CursorTokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// CursorAuthBundle contains the complete authentication data.
type CursorAuthBundle struct {
	TokenData CursorTokenData
	UserInfo  string
}

// CursorTokenStorage represents the storage format for Cursor tokens.
type CursorTokenStorage struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	UserInfo     string `json:"user_info"`
	Type         string `json:"type"`
}

// DeviceCodeResponse represents the response from the device code endpoint.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

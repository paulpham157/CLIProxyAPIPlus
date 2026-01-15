package cursor

import (
	"fmt"
)

// ErrorType represents the type of authentication error.
type ErrorType string

const (
	// ErrTypeDeviceCode indicates a failure in device code request.
	ErrTypeDeviceCode ErrorType = "device_code_failed"
	// ErrTypeTokenExchange indicates a failure in token exchange.
	ErrTypeTokenExchange ErrorType = "token_exchange_failed"
	// ErrTypeAuthPending indicates authorization is pending.
	ErrTypeAuthPending ErrorType = "authorization_pending"
	// ErrTypeSlowDown indicates the client should slow down polling.
	ErrTypeSlowDown ErrorType = "slow_down"
	// ErrTypeExpired indicates the device code expired.
	ErrTypeExpired ErrorType = "expired_token"
	// ErrTypeAccessDenied indicates the user denied access.
	ErrTypeAccessDenied ErrorType = "access_denied"
	// ErrTypePollingTimeout indicates polling timed out.
	ErrTypePollingTimeout ErrorType = "polling_timeout"
	// ErrTypeUserInfo indicates a failure in user info retrieval.
	ErrTypeUserInfo ErrorType = "user_info_failed"
)

// AuthenticationError represents an authentication-specific error.
type AuthenticationError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *AuthenticationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("cursor auth: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("cursor auth: %s", e.Message)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

// NewAuthenticationError creates a new authentication error.
func NewAuthenticationError(baseErr *AuthenticationError, err error) *AuthenticationError {
	if baseErr == nil {
		return &AuthenticationError{
			Type:    ErrTypeTokenExchange,
			Message: "unknown error",
			Err:     err,
		}
	}
	return &AuthenticationError{
		Type:    baseErr.Type,
		Message: baseErr.Message,
		Err:     err,
	}
}

// Predefined authentication errors.
var (
	ErrDeviceCodeFailed = &AuthenticationError{
		Type:    ErrTypeDeviceCode,
		Message: "failed to request device code",
	}
	ErrTokenExchangeFailed = &AuthenticationError{
		Type:    ErrTypeTokenExchange,
		Message: "failed to exchange token",
	}
	ErrAuthorizationPending = &AuthenticationError{
		Type:    ErrTypeAuthPending,
		Message: "authorization pending",
	}
	ErrSlowDown = &AuthenticationError{
		Type:    ErrTypeSlowDown,
		Message: "slow down polling",
	}
	ErrDeviceCodeExpired = &AuthenticationError{
		Type:    ErrTypeExpired,
		Message: "device code expired",
	}
	ErrAccessDenied = &AuthenticationError{
		Type:    ErrTypeAccessDenied,
		Message: "access denied by user",
	}
	ErrPollingTimeout = &AuthenticationError{
		Type:    ErrTypePollingTimeout,
		Message: "polling timed out",
	}
	ErrUserInfoFailed = &AuthenticationError{
		Type:    ErrTypeUserInfo,
		Message: "failed to fetch user info",
	}
)

// OAuthError represents an OAuth-specific error from the Cursor API.
type OAuthError struct {
	Error            string
	ErrorDescription string
	StatusCode       int
}

func (e *OAuthError) Error() string {
	if e.ErrorDescription != "" {
		return fmt.Sprintf("cursor oauth: %s: %s (status %d)", e.Error, e.ErrorDescription, e.StatusCode)
	}
	return fmt.Sprintf("cursor oauth: %s (status %d)", e.Error, e.StatusCode)
}

// NewOAuthError creates a new OAuth error.
func NewOAuthError(error, description string, statusCode int) *OAuthError {
	return &OAuthError{
		Error:            error,
		ErrorDescription: description,
		StatusCode:       statusCode,
	}
}

// GetUserFriendlyMessage returns a user-friendly error message.
func GetUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	var authErr *AuthenticationError
	if e, ok := err.(*AuthenticationError); ok {
		authErr = e
	}

	if authErr != nil {
		switch authErr.Type {
		case ErrTypeDeviceCode:
			return "Failed to start authentication. Please check your internet connection and try again."
		case ErrTypeTokenExchange:
			return "Failed to complete authentication. Please try again."
		case ErrTypeExpired:
			return "The authentication code has expired. Please start the authentication process again."
		case ErrTypeAccessDenied:
			return "Authentication was denied. Please try again and authorize the application."
		case ErrTypePollingTimeout:
			return "Authentication timed out. Please start the authentication process again and complete it more quickly."
		case ErrTypeUserInfo:
			return "Failed to retrieve user information. Your authentication may still be valid."
		}
	}

	return fmt.Sprintf("Authentication failed: %v", err)
}

# Continue.dev Provider Implementation

This document describes the implementation of Continue.dev provider integration for CLIProxyAPI Plus.

## Overview

The Continue.dev provider integration enables users to authenticate with Continue.dev using OAuth device flow and access various AI models through the Continue.dev API. The implementation follows the existing patterns in CLIProxyAPI Plus for provider integration.

## Components Implemented

### 1. Authentication Module (`internal/auth/continue/`)

#### Files Created:
- **`continue_auth.go`**: Core authentication service
  - `ContinueAuth`: Main authentication service struct
  - `StartDeviceFlow()`: Initiates OAuth device flow
  - `WaitForAuthorization()`: Polls for user authorization
  - `GetContinueAPIToken()`: Exchanges access token for API token
  - `ValidateToken()`: Validates existing tokens
  - `CreateTokenStorage()`: Creates token storage structure

- **`oauth.go`**: OAuth device flow implementation
  - `DeviceFlowClient`: Handles device flow operations
  - `RequestDeviceCode()`: Requests device code from Continue.dev
  - `PollForToken()`: Polls for token exchange
  - `exchangeDeviceCode()`: Exchanges device code for access token
  - `FetchUserInfo()`: Retrieves user information

- **`token.go`**: Token storage structures
  - `ContinueTokenStorage`: Stores OAuth tokens persistently
  - `ContinueTokenData`: Raw OAuth token response
  - `ContinueAuthBundle`: Bundles auth data
  - `DeviceCodeResponse`: Device code flow response

- **`errors.go`**: Error handling
  - `OAuthError`: OAuth-specific errors
  - `AuthenticationError`: General authentication errors
  - Standard OAuth error types (device_code_failed, authorization_pending, etc.)
  - `GetUserFriendlyMessage()`: User-friendly error messages

### 2. SDK Authentication (`sdk/auth/continue.go`)

- **`ContinueAuthenticator`**: Implements the `Authenticator` interface
  - `Provider()`: Returns "continue"
  - `Login()`: Performs device flow authentication with browser auto-open
  - `RefreshLead()`: Returns nil (tokens don't need refresh)
- **`RefreshContinueToken()`**: Token validation function

### 3. Runtime Executor (`internal/runtime/executor/continue_executor.go`)

- **`ContinueExecutor`**: Handles API requests to Continue.dev
  - `Execute()`: Non-streaming requests
  - `ExecuteStream()`: Streaming requests
  - `PrepareRequest()`: Adds authentication headers
  - `HttpRequest()`: Executes HTTP requests with authentication
  - Token caching with expiry management
  - Request/response logging and usage tracking

### 4. Translators

#### OpenAI ↔ Continue Translators (`internal/translator/openai/continue/`)
- **`openai_continue_request.go`**: OpenAI to Continue request translation
- **`openai_continue_response.go`**: Continue to OpenAI response translation
- **`init.go`**: Registers OpenAI→Continue translators

#### Continue ↔ OpenAI Translators (`internal/translator/continue/openai/chat-completions/`)
- **`continue_openai_request.go`**: Continue to OpenAI request translation
- **`continue_openai_response.go`**: OpenAI to Continue response translation
- **`init.go`**: Registers Continue→OpenAI translators

#### Claude ↔ Continue Translators
- `internal/translator/claude/continue/`: Claude to Continue translation
- `internal/translator/continue/claude/`: Continue to Claude translation

### 5. Model Registry (`internal/registry/model_definitions.go`)

Added `GetContinueModels()` function that returns available Continue.dev models:
- gpt-4o
- gpt-4o-mini
- claude-3-5-sonnet-20241022
- claude-3-5-haiku-20241022
- gemini-2.0-flash-exp
- gemini-1.5-pro
- deepseek-chat
- deepseek-coder

### 6. Constants and Formats

- **`internal/constant/constant.go`**: Added `Continue = "continue"`
- **`sdk/translator/formats.go`**: Added `FormatContinue Format = "continue"`

### 7. Service Integration (`sdk/cliproxy/service.go`)

Registered Continue.dev in two locations:
1. **Executor registration** (line 388-389):
   ```go
   case "continue":
       s.coreManager.RegisterExecutor(executor.NewContinueExecutor(s.cfg))
   ```

2. **Model registration** (line 786-788):
   ```go
   case "continue":
       models = registry.GetContinueModels()
       models = applyExcludedModels(models, excluded)
   ```

### 8. Authentication Registry (`sdk/auth/refresh_registry.go`)

Registered Continue authenticator:
```go
registerRefreshLead("continue", func() Authenticator { return NewContinueAuthenticator() })
```

### 9. Translator Registration (`internal/translator/init.go`)

Registered all Continue translator packages:
```go
_ "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/continue/openai/chat-completions"
_ "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/openai/continue"
_ "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/continue/claude"
_ "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/claude/continue"
```

## API Endpoints

The Continue.dev implementation uses the following endpoints:

### Authentication:
- **Device Code**: `https://auth.continue.dev/device/code`
- **Token Exchange**: `https://auth.continue.dev/oauth/token`
- **User Info**: `https://api.continue.dev/user`
- **API Token**: `https://api.continue.dev/auth/token`

### API:
- **Base URL**: `https://api.continue.dev`
- **Chat Completions**: `/v1/chat/completions`

## OAuth Device Flow

1. **Initiate Flow**: Request device code with client ID
2. **Display Code**: Show verification URL and user code
3. **Poll for Token**: Poll token endpoint until user authorizes
4. **Exchange Token**: Get access token after authorization
5. **Get API Token**: Exchange access token for Continue.dev API token
6. **Cache Token**: Cache API token with expiry

## Usage

### Authentication

```bash
# CLI authentication (not shown in code, but would be used)
cli-proxy-api auth login continue
```

The authentication flow will:
1. Display a verification URL (e.g., `https://auth.continue.dev/activate`)
2. Show a user code to enter
3. Automatically open browser (unless `--no-browser` flag is used)
4. Wait for user authorization (15 minute timeout)
5. Validate Continue.dev access
6. Save credentials to `auths/continue-{username}.json`

### Model Access

After authentication, Continue.dev models are automatically available via:
- `/v1/models` endpoint (lists all available models)
- `/v1/chat/completions` endpoint (for chat requests)

## Technical Features

### Token Management
- **Caching**: API tokens cached for 25 minutes
- **Expiry Buffer**: 5-minute buffer before token expiry
- **Auto-refresh**: Tokens refreshed when within expiry buffer

### Request Translation
- **Pass-through**: Continue.dev uses OpenAI-compatible format
- **Model normalization**: Model names preserved in requests
- **Payload configuration**: Supports config-based payload modifications

### Streaming Support
- **SSE Parsing**: Server-Sent Events with 20MB buffer size
- **Chunk Processing**: Line-by-line streaming response handling
- **Context Cancellation**: Graceful cancellation support

### Error Handling
- **OAuth Errors**: Standard OAuth error codes (authorization_pending, slow_down, etc.)
- **User-Friendly Messages**: Translated error messages for better UX
- **Retry Logic**: Built-in retry for transient errors (slow_down)

## Security Considerations

1. **Token Storage**: Tokens stored in JSON files with 0700 directory permissions
2. **No Secrets in Code**: Client ID is public (device flow design)
3. **Token Expiry**: Tokens have expiration and are refreshed automatically
4. **HTTPS Only**: All API communication over HTTPS

## Integration with Existing Systems

The Continue.dev provider integrates seamlessly with:
- **Model Registry**: Models exposed via `/v1/models` endpoint
- **Usage Tracking**: Request/response logging and token usage tracking
- **Proxy Support**: Respects proxy configuration from main config
- **Authentication Manager**: Uses standard auth storage and management

## File Structure

```
internal/
  auth/
    continue/
      continue_auth.go
      oauth.go
      token.go
      errors.go
  constant/
    constant.go (modified)
  registry/
    model_definitions.go (modified)
  runtime/
    executor/
      continue_executor.go
  translator/
    continue/
      openai/
        chat-completions/
          continue_openai_request.go
          continue_openai_response.go
          init.go
      claude/
        continue_claude_request.go
        continue_claude_response.go
        init.go
    openai/
      continue/
        openai_continue_request.go
        openai_continue_response.go
        init.go
    claude/
      continue/
        claude_continue_request.go
        claude_continue_response.go
        init.go
    init.go (modified)

sdk/
  auth/
    continue.go
    refresh_registry.go (modified)
  cliproxy/
    service.go (modified)
  translator/
    formats.go (modified)
```

## Testing Recommendations

To test the implementation:

1. **Authentication Flow**:
   - Test successful authentication
   - Test browser auto-open (with and without --no-browser)
   - Test timeout scenarios
   - Test user denial

2. **API Requests**:
   - Test non-streaming requests
   - Test streaming requests
   - Test various models
   - Test error handling

3. **Token Management**:
   - Test token caching
   - Test token expiry and refresh
   - Test invalid tokens

4. **Model Listing**:
   - Verify models appear in `/v1/models`
   - Verify model metadata is correct

## Future Enhancements

Potential improvements:
1. Add model-specific configurations
2. Implement rate limiting
3. Add model capability detection
4. Support additional Continue.dev features
5. Add telemetry and monitoring

## Compliance

This implementation:
- Follows CLIProxyAPI Plus architecture patterns
- Uses existing authentication interfaces
- Integrates with model registry system
- Supports standard API endpoints
- Maintains backward compatibility

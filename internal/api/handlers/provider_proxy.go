// Package handlers provides HTTP handlers for the API server.
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/interfaces"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers"
	"github.com/tidwall/gjson"
	log "github.com/sirupsen/logrus"
)

// ProviderProxyHandler handles POST requests to /api/providers/:provider
// It validates the provider type, forwards to the appropriate provider service,
// and streams responses back using SSE format.
type ProviderProxyHandler struct {
	baseHandler *handlers.BaseAPIHandler
}

// NewProviderProxyHandler creates a new provider proxy handler instance.
func NewProviderProxyHandler(baseHandler *handlers.BaseAPIHandler) *ProviderProxyHandler {
	return &ProviderProxyHandler{
		baseHandler: baseHandler,
	}
}

// validProviders defines the set of supported provider types
var validProviders = map[string]bool{
	"openai":        true,
	"claude":        true,
	"anthropic":     true,
	"gemini":        true,
	"codex":         true,
	"kiro":          true,
	"copilot":       true,
	"github":        true,
	"antigravity":   true,
	"qwen":          true,
	"vertex":        true,
	"gemini-cli":    true,
	"openai-compat": true,
}

// HandleProviderProxy handles POST /api/providers/:provider requests
func (h *ProviderProxyHandler) HandleProviderProxy(c *gin.Context) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))

	// Validate provider type
	if provider == "" {
		c.JSON(http.StatusBadRequest, handlers.ErrorResponse{
			Error: handlers.ErrorDetail{
				Message: "Provider parameter is required",
				Type:    "invalid_request_error",
				Code:    "missing_provider",
			},
		})
		return
	}

	if !validProviders[provider] {
		c.JSON(http.StatusBadRequest, handlers.ErrorResponse{
			Error: handlers.ErrorDetail{
				Message: fmt.Sprintf("Unsupported provider: %s", provider),
				Type:    "invalid_request_error",
				Code:    "invalid_provider",
			},
		})
		return
	}

	// Read request body
	rawJSON, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, handlers.ErrorResponse{
			Error: handlers.ErrorDetail{
				Message: fmt.Sprintf("Failed to read request body: %v", err),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// Parse and validate JSON
	if !json.Valid(rawJSON) {
		c.JSON(http.StatusBadRequest, handlers.ErrorResponse{
			Error: handlers.ErrorDetail{
				Message: "Invalid JSON in request body",
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// Extract model from request
	modelName := gjson.GetBytes(rawJSON, "model").String()
	if modelName == "" {
		c.JSON(http.StatusBadRequest, handlers.ErrorResponse{
			Error: handlers.ErrorDetail{
				Message: "Model parameter is required",
				Type:    "invalid_request_error",
				Code:    "missing_model",
			},
		})
		return
	}

	// Check if streaming is requested
	streamResult := gjson.GetBytes(rawJSON, "stream")
	stream := streamResult.Type == gjson.True

	// Get alt parameter (for Gemini compatibility)
	alt := h.baseHandler.GetAlt(c)

	// Determine handler type based on provider
	handlerType := h.getHandlerType(provider)

	if stream {
		h.handleStreamingResponse(c, handlerType, modelName, rawJSON, alt)
	} else {
		h.handleNonStreamingResponse(c, handlerType, modelName, rawJSON, alt)
	}
}

// getHandlerType maps provider names to handler types
func (h *ProviderProxyHandler) getHandlerType(provider string) string {
	switch provider {
	case "openai", "openai-compat":
		return "openai"
	case "claude", "anthropic":
		return "claude"
	case "gemini", "vertex":
		return "gemini"
	case "gemini-cli":
		return "gemini-cli"
	case "codex":
		return "codex"
	case "kiro":
		return "kiro"
	case "copilot", "github":
		return "copilot"
	case "antigravity":
		return "antigravity"
	case "qwen":
		return "qwen"
	default:
		return "openai" // Default to OpenAI format
	}
}

// handleStreamingResponse processes streaming requests and forwards SSE responses
func (h *ProviderProxyHandler) handleStreamingResponse(c *gin.Context, handlerType, modelName string, rawJSON []byte, alt string) {
	ctx, cancel := h.baseHandler.GetContextWithCancel(h.baseHandler, c, c.Request.Context())

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.baseHandler.WriteErrorResponse(c, &interfaces.ErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Error:      fmt.Errorf("streaming not supported"),
		})
		cancel()
		return
	}

	// Execute streaming request
	data, errs := h.baseHandler.ExecuteStreamWithAuthManager(ctx, handlerType, modelName, rawJSON, alt)

	// Forward stream to client
	h.baseHandler.ForwardStream(c, flusher, cancel, data, errs, handlers.StreamForwardOptions{
		WriteChunk: func(chunk []byte) {
			// Write in SSE format: "data: {json}\n\n"
			if len(chunk) > 0 {
				_, _ = c.Writer.Write([]byte("data: "))
				_, _ = c.Writer.Write(chunk)
				_, _ = c.Writer.Write([]byte("\n\n"))
			}
		},
		WriteTerminalError: func(errMsg *interfaces.ErrorMessage) {
			// Write error in SSE format
			status := http.StatusInternalServerError
			if errMsg != nil && errMsg.StatusCode > 0 {
				status = errMsg.StatusCode
			}
			errText := http.StatusText(status)
			if errMsg != nil && errMsg.Error != nil {
				errText = errMsg.Error.Error()
			}
			errorBody := handlers.BuildErrorResponseBody(status, errText)
			_, _ = c.Writer.Write([]byte("data: "))
			_, _ = c.Writer.Write(errorBody)
			_, _ = c.Writer.Write([]byte("\n\n"))
		},
		WriteDone: func() {
			// Write [DONE] marker
			_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
		},
		WriteKeepAlive: func() {
			// Write SSE comment as keep-alive
			_, _ = c.Writer.Write([]byte(": keep-alive\n\n"))
		},
	})
}

// handleNonStreamingResponse processes non-streaming requests
func (h *ProviderProxyHandler) handleNonStreamingResponse(c *gin.Context, handlerType, modelName string, rawJSON []byte, alt string) {
	ctx, cancel := h.baseHandler.GetContextWithCancel(h.baseHandler, c, c.Request.Context())
	defer cancel()

	// Execute non-streaming request
	response, errMsg := h.baseHandler.ExecuteWithAuthManager(ctx, handlerType, modelName, rawJSON, alt)

	if errMsg != nil {
		log.Errorf("Provider proxy error: %v", errMsg.Error)
		h.baseHandler.WriteErrorResponse(c, errMsg)
		return
	}

	// Write successful response
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write(response)
}

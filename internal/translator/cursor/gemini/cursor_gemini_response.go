package gemini

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini/openai/chat-completions"
)

// ConvertCursorResponseToGemini converts Cursor streaming response to Gemini format.
// Uses the existing OpenAI to Gemini translator since Cursor is OpenAI-compatible.
func ConvertCursorResponseToGemini(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return openaitranslator.ConvertOpenAIResponseToGemini(ctx, originalRequest, translatedRequest, model, line, param)
}

// ConvertCursorResponseToGeminiNonStream converts Cursor non-streaming response to Gemini format.
// Uses the existing OpenAI to Gemini translator since Cursor is OpenAI-compatible.
func ConvertCursorResponseToGeminiNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return openaitranslator.ConvertOpenAIResponseToGeminiNonStream(ctx, originalRequest, translatedRequest, model, body, param)
}

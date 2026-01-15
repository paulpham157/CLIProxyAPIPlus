package gemini

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini/openai/chat-completions"
)

// ConvertWindsurfResponseToGemini converts Windsurf streaming response to Gemini format.
// Uses the existing OpenAI to Gemini translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfResponseToGemini(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return openaitranslator.ConvertOpenAIResponseToGemini(ctx, originalRequest, translatedRequest, model, line, param)
}

// ConvertWindsurfResponseToGeminiNonStream converts Windsurf non-streaming response to Gemini format.
// Uses the existing OpenAI to Gemini translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfResponseToGeminiNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return openaitranslator.ConvertOpenAIResponseToGeminiNonStream(ctx, originalRequest, translatedRequest, model, body, param)
}

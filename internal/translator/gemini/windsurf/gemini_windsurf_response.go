package windsurf

import (
	"context"

	geminitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini/openai/chat-completions"
)

// ConvertGeminiResponseToWindsurf converts Gemini streaming response to Windsurf format.
// Uses the existing Gemini to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertGeminiResponseToWindsurf(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return geminitranslator.ConvertGeminiResponseToOpenAI(ctx, originalRequest, translatedRequest, model, line, param.(*any))
}

// ConvertGeminiResponseToWindsurfNonStream converts Gemini non-streaming response to Windsurf format.
// Uses the existing Gemini to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertGeminiResponseToWindsurfNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return geminitranslator.ConvertGeminiResponseToOpenAINonStream(ctx, originalRequest, translatedRequest, model, body, param.(*any))
}

// WindsurfTokenCount converts token count response to Windsurf format.
func WindsurfTokenCount(ctx context.Context, count int64) string {
	return geminitranslator.OpenAITokenCount(ctx, count)
}

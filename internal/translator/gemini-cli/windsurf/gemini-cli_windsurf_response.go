package windsurf

import (
	"context"

	geminitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini-cli/openai"
)

// ConvertGeminiCLIResponseToWindsurf converts Gemini-CLI streaming response to Windsurf format.
// Uses the existing Gemini-CLI to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertGeminiCLIResponseToWindsurf(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return geminitranslator.ConvertGeminiCLIResponseToOpenAI(ctx, originalRequest, translatedRequest, model, line, param.(*any))
}

// ConvertGeminiCLIResponseToWindsurfNonStream converts Gemini-CLI non-streaming response to Windsurf format.
// Uses the existing Gemini-CLI to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertGeminiCLIResponseToWindsurfNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return geminitranslator.ConvertGeminiCLIResponseToOpenAINonStream(ctx, originalRequest, translatedRequest, model, body, param.(*any))
}

// WindsurfTokenCount converts token count response to Windsurf format.
func WindsurfTokenCount(ctx context.Context, count int64) string {
	return geminitranslator.OpenAITokenCount(ctx, count)
}

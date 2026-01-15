package geminiCLI

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/openai/gemini-cli"
)

// ConvertWindsurfResponseToGeminiCLI converts Windsurf streaming response to Gemini-CLI format.
// Uses the existing OpenAI to Gemini-CLI translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfResponseToGeminiCLI(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return openaitranslator.ConvertOpenAIResponseToGeminiCLI(ctx, model, originalRequest, translatedRequest, line, param.(*any))
}

// ConvertWindsurfResponseToGeminiCLINonStream converts Windsurf non-streaming response to Gemini-CLI format.
// Uses the existing OpenAI to Gemini-CLI translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfResponseToGeminiCLINonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return openaitranslator.ConvertOpenAIResponseToGeminiCLINonStream(ctx, model, originalRequest, translatedRequest, body, param.(*any))
}

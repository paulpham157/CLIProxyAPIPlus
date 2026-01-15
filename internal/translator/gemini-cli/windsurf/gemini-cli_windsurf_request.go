package windsurf

import (
	"context"

	geminitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini-cli/openai"
)

// ConvertWindsurfRequestToGeminiCLI converts Windsurf request format to Gemini-CLI format.
// Uses the existing OpenAI to Gemini-CLI translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfRequestToGeminiCLI(ctx context.Context, model string, body []byte, stream bool) []byte {
	return geminitranslator.ConvertOpenAIRequestToGeminiCLI(ctx, model, body, stream)
}

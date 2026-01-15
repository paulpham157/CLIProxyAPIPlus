package geminiCLI

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/openai/gemini-cli"
)

// ConvertGeminiCLIRequestToWindsurf converts Gemini-CLI request format to Windsurf format.
// Uses the existing Gemini-CLI to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertGeminiCLIRequestToWindsurf(ctx context.Context, model string, body []byte, stream bool) []byte {
	return openaitranslator.ConvertGeminiCLIRequestToOpenAI(model, body, stream)
}

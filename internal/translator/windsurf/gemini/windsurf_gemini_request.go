package gemini

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini/openai/chat-completions"
)

// ConvertGeminiRequestToWindsurf converts Gemini request format to Windsurf format.
// Uses the existing Gemini to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertGeminiRequestToWindsurf(ctx context.Context, model string, body []byte, stream bool) []byte {
	return openaitranslator.ConvertGeminiRequestToOpenAI(ctx, model, body, stream)
}

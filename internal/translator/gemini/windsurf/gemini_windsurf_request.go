package windsurf

import (
	"context"

	geminitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini/openai/chat-completions"
)

// ConvertWindsurfRequestToGemini converts Windsurf request format to Gemini format.
// Uses the existing OpenAI to Gemini translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfRequestToGemini(ctx context.Context, model string, body []byte, stream bool) []byte {
	return geminitranslator.ConvertOpenAIRequestToGemini(ctx, model, body, stream)
}

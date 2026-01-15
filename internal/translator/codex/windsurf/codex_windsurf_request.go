package windsurf

import (
	"context"

	codextranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/codex/openai/chat-completions"
)

// ConvertWindsurfRequestToCodex converts Windsurf request format to Codex format.
// Uses the existing OpenAI to Codex translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfRequestToCodex(ctx context.Context, model string, body []byte, stream bool) []byte {
	return codextranslator.ConvertOpenAIRequestToCodex(ctx, model, body, stream)
}

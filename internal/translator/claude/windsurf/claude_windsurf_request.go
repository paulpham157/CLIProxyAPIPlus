package windsurf

import (
	"context"

	claudetranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/claude/openai/chat-completions"
)

// ConvertWindsurfRequestToClaude converts Windsurf request format to Claude format.
// Uses the existing OpenAI to Claude translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfRequestToClaude(ctx context.Context, model string, body []byte, stream bool) []byte {
	return claudetranslator.ConvertOpenAIRequestToClaude(ctx, model, body, stream)
}

package claude

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/claude/openai/chat-completions"
)

// ConvertClaudeRequestToWindsurf converts Claude request format to Windsurf format.
// Uses the existing Claude to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertClaudeRequestToWindsurf(ctx context.Context, model string, body []byte, stream bool) []byte {
	return openaitranslator.ConvertClaudeRequestToOpenAI(ctx, model, body, stream)
}

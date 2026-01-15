package claude

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/claude/openai/chat-completions"
)

// ConvertClaudeRequestToCursor converts Claude request format to Cursor format.
// Uses the existing Claude to OpenAI translator since Cursor is OpenAI-compatible.
func ConvertClaudeRequestToCursor(ctx context.Context, model string, body []byte, stream bool) []byte {
	return openaitranslator.ConvertClaudeRequestToOpenAI(ctx, model, body, stream)
}

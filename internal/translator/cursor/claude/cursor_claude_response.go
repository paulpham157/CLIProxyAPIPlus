package claude

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/claude/openai/chat-completions"
)

// ConvertCursorResponseToClaude converts Cursor streaming response to Claude format.
// Uses the existing OpenAI to Claude translator since Cursor is OpenAI-compatible.
func ConvertCursorResponseToClaude(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return openaitranslator.ConvertOpenAIResponseToClaude(ctx, originalRequest, translatedRequest, model, line, param)
}

// ConvertCursorResponseToClaudeNonStream converts Cursor non-streaming response to Claude format.
// Uses the existing OpenAI to Claude translator since Cursor is OpenAI-compatible.
func ConvertCursorResponseToClaudeNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return openaitranslator.ConvertOpenAIResponseToClaudeNonStream(ctx, originalRequest, translatedRequest, model, body, param)
}

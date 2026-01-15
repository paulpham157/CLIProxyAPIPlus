package windsurf

import (
	"context"

	claudetranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/claude/openai/chat-completions"
)

// ConvertClaudeResponseToWindsurf converts Claude streaming response to Windsurf format.
// Uses the existing Claude to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertClaudeResponseToWindsurf(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return claudetranslator.ConvertClaudeResponseToOpenAI(ctx, originalRequest, translatedRequest, model, line, param.(*any))
}

// ConvertClaudeResponseToWindsurfNonStream converts Claude non-streaming response to Windsurf format.
// Uses the existing Claude to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertClaudeResponseToWindsurfNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return claudetranslator.ConvertClaudeResponseToOpenAINonStream(ctx, originalRequest, translatedRequest, model, body, param.(*any))
}

// WindsurfTokenCount converts token count response to Windsurf format.
func WindsurfTokenCount(ctx context.Context, count int64) string {
	return claudetranslator.OpenAITokenCount(ctx, count)
}

package claude

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/claude/openai/chat-completions"
)

// ConvertWindsurfResponseToClaude converts Windsurf streaming response to Claude format.
// Uses the existing OpenAI to Claude translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfResponseToClaude(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return openaitranslator.ConvertOpenAIResponseToClaude(ctx, originalRequest, translatedRequest, model, line, param)
}

// ConvertWindsurfResponseToClaudeNonStream converts Windsurf non-streaming response to Claude format.
// Uses the existing OpenAI to Claude translator since Windsurf is OpenAI-compatible.
func ConvertWindsurfResponseToClaudeNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return openaitranslator.ConvertOpenAIResponseToClaudeNonStream(ctx, originalRequest, translatedRequest, model, body, param)
}

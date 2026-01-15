package windsurf

import (
	"context"

	codextranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/codex/openai/responses"
)

// ConvertCodexResponseToWindsurf converts Codex streaming response to Windsurf format.
// Uses the existing Codex to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertCodexResponseToWindsurf(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	return codextranslator.ConvertCodexResponseToOpenAIResponse(ctx, originalRequest, translatedRequest, model, line, param)
}

// ConvertCodexResponseToWindsurfNonStream converts Codex non-streaming response to Windsurf format.
// Uses the existing Codex to OpenAI translator since Windsurf is OpenAI-compatible.
func ConvertCodexResponseToWindsurfNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return codextranslator.ConvertCodexResponseToOpenAIResponseNonStream(ctx, originalRequest, translatedRequest, model, body, param)
}

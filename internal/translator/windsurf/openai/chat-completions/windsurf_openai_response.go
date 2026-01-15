package chat_completions

import (
	"context"
)

// ConvertWindsurfResponseToOpenAI converts Windsurf streaming response to OpenAI format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertWindsurfResponseToOpenAI(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	if len(line) == 0 {
		return nil
	}
	return []string{string(line)}
}

// ConvertWindsurfResponseToOpenAINonStream converts Windsurf non-streaming response to OpenAI format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertWindsurfResponseToOpenAINonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return string(body)
}

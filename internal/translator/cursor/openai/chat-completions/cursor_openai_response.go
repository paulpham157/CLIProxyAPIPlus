package chat_completions

import (
	"context"
)

// ConvertCursorResponseToOpenAI converts Cursor streaming response to OpenAI format.
// Since Cursor uses OpenAI-compatible API, this is a pass-through.
func ConvertCursorResponseToOpenAI(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	if len(line) == 0 {
		return nil
	}
	return []string{string(line)}
}

// ConvertCursorResponseToOpenAINonStream converts Cursor non-streaming response to OpenAI format.
// Since Cursor uses OpenAI-compatible API, this is a pass-through.
func ConvertCursorResponseToOpenAINonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return string(body)
}

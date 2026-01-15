package responses

import (
	"context"
)

// ConvertCursorResponseToOpenAIResponse converts Cursor streaming response to OpenAI response format.
// Since Cursor uses OpenAI-compatible API, this is a pass-through.
func ConvertCursorResponseToOpenAIResponse(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	if len(line) == 0 {
		return nil
	}
	return []string{string(line)}
}

// ConvertCursorResponseToOpenAIResponseNonStream converts Cursor non-streaming response to OpenAI response format.
// Since Cursor uses OpenAI-compatible API, this is a pass-through.
func ConvertCursorResponseToOpenAIResponseNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return string(body)
}

package chat_completions

import (
	"context"
)

// ConvertOpenAIRequestToCursor converts OpenAI request format to Cursor format.
// Since Cursor uses OpenAI-compatible API, this is a pass-through.
func ConvertOpenAIRequestToCursor(ctx context.Context, model string, body []byte, stream bool) []byte {
	return body
}

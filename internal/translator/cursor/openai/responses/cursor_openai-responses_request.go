package responses

import (
	"context"
)

// ConvertOpenAIResponseRequestToCursor converts OpenAI response format request to Cursor format.
// Since Cursor uses OpenAI-compatible API, this is a pass-through.
func ConvertOpenAIResponseRequestToCursor(ctx context.Context, model string, body []byte, stream bool) []byte {
	return body
}

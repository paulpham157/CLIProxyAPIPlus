package responses

import (
	"context"
)

// ConvertOpenAIResponseRequestToWindsurf converts OpenAI response format request to Windsurf format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertOpenAIResponseRequestToWindsurf(ctx context.Context, model string, body []byte, stream bool) []byte {
	return body
}

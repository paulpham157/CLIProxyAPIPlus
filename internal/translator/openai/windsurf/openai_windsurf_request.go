package windsurf

import (
	"context"
)

// ConvertWindsurfRequestToOpenAI converts Windsurf request format to OpenAI format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertWindsurfRequestToOpenAI(ctx context.Context, model string, body []byte, stream bool) []byte {
	return body
}

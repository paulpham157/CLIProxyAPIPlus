package chat_completions

import (
	"context"
)

// ConvertOpenAIRequestToWindsurf converts OpenAI request format to Windsurf format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertOpenAIRequestToWindsurf(ctx context.Context, model string, body []byte, stream bool) []byte {
	return body
}

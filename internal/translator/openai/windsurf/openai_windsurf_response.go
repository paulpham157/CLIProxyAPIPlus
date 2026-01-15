package windsurf

import (
	"context"
)

// ConvertOpenAIResponseToWindsurf converts OpenAI streaming response to Windsurf format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertOpenAIResponseToWindsurf(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	if len(line) == 0 {
		return nil
	}
	return []string{string(line)}
}

// ConvertOpenAIResponseToWindsurfNonStream converts OpenAI non-streaming response to Windsurf format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertOpenAIResponseToWindsurfNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return string(body)
}

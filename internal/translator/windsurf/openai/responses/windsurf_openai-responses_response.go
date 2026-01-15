package responses

import (
	"context"
)

// ConvertWindsurfResponseToOpenAIResponse converts Windsurf streaming response to OpenAI response format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertWindsurfResponseToOpenAIResponse(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, line []byte, param any) []string {
	if len(line) == 0 {
		return nil
	}
	return []string{string(line)}
}

// ConvertWindsurfResponseToOpenAIResponseNonStream converts Windsurf non-streaming response to OpenAI response format.
// Since Windsurf uses OpenAI-compatible API, this is a pass-through.
func ConvertWindsurfResponseToOpenAIResponseNonStream(ctx context.Context, originalRequest []byte, translatedRequest []byte, model string, body []byte, param any) string {
	return string(body)
}

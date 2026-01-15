package chatcompletions

import (
	"context"
)

func ConvertOpenAIResponseToContinue(_ context.Context, _ string, originalRequestRawJSON, requestRawJSON, rawJSON []byte, param *any) []string {
	return []string{string(rawJSON)}
}

func ConvertOpenAIResponseToContinueNonStream(_ context.Context, _ string, originalRequestRawJSON, requestRawJSON, rawJSON []byte, _ *any) string {
	return string(rawJSON)
}

func OpenAITokenCount(ctx context.Context, count int64) string {
	return ""
}

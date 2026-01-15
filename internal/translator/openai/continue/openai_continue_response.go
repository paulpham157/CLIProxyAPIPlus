package continuetr

import (
	"context"
)

func ConvertContinueResponseToOpenAI(_ context.Context, _ string, originalRequestRawJSON, requestRawJSON, rawJSON []byte, param *any) []string {
	return []string{string(rawJSON)}
}

func ConvertContinueResponseToOpenAINonStream(_ context.Context, _ string, originalRequestRawJSON, requestRawJSON, rawJSON []byte, _ *any) string {
	return string(rawJSON)
}

func ContinueTokenCount(ctx context.Context, count int64) string {
	return ""
}

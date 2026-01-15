package claude

import (
	"context"
)

func ConvertClaudeResponseToContinue(_ context.Context, _ string, originalRequestRawJSON, requestRawJSON, rawJSON []byte, param *any) []string {
	return []string{string(rawJSON)}
}

func ConvertClaudeResponseToContinueNonStream(_ context.Context, _ string, originalRequestRawJSON, requestRawJSON, rawJSON []byte, _ *any) string {
	return string(rawJSON)
}

func ClaudeTokenCount(ctx context.Context, count int64) string {
	return ""
}

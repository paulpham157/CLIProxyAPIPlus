package gemini

import (
	"context"

	openaitranslator "github.com/router-for-me/CLIProxyAPI/v6/internal/translator/gemini/openai/chat-completions"
)

// ConvertGeminiRequestToCursor converts Gemini request format to Cursor format.
// Uses the existing Gemini to OpenAI translator since Cursor is OpenAI-compatible.
func ConvertGeminiRequestToCursor(ctx context.Context, model string, body []byte, stream bool) []byte {
	return openaitranslator.ConvertGeminiRequestToOpenAI(ctx, model, body, stream)
}

package chatcompletions

import (
	. "github.com/router-for-me/CLIProxyAPI/v6/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/interfaces"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		Continue,
		OpenAI,
		ConvertContinueRequestToOpenAI,
		interfaces.TranslateResponse{
			Stream:     ConvertOpenAIResponseToContinue,
			NonStream:  ConvertOpenAIResponseToContinueNonStream,
			TokenCount: OpenAITokenCount,
		},
	)
}

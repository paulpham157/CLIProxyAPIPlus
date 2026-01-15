package continuetr

import (
	. "github.com/router-for-me/CLIProxyAPI/v6/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/interfaces"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		OpenAI,
		Continue,
		ConvertOpenAIRequestToContinue,
		interfaces.TranslateResponse{
			Stream:     ConvertContinueResponseToOpenAI,
			NonStream:  ConvertContinueResponseToOpenAINonStream,
			TokenCount: ContinueTokenCount,
		},
	)
}

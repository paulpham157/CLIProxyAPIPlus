package chat_completions

import (
	. "github.com/router-for-me/CLIProxyAPI/v6/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/interfaces"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		OpenAI,
		Windsurf,
		ConvertOpenAIRequestToWindsurf,
		interfaces.TranslateResponse{
			Stream:    ConvertWindsurfResponseToOpenAI,
			NonStream: ConvertWindsurfResponseToOpenAINonStream,
		},
	)
}

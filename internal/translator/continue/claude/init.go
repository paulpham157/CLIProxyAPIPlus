package claude

import (
	. "github.com/router-for-me/CLIProxyAPI/v6/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/interfaces"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		Continue,
		Claude,
		ConvertContinueRequestToClaude,
		interfaces.TranslateResponse{
			Stream:     ConvertClaudeResponseToContinue,
			NonStream:  ConvertClaudeResponseToContinueNonStream,
			TokenCount: ClaudeTokenCount,
		},
	)
}

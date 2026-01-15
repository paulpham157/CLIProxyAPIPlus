package windsurf

import (
	. "github.com/router-for-me/CLIProxyAPI/v6/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/interfaces"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/translator/translator"
)

func init() {
	translator.Register(
		Windsurf,
		Claude,
		ConvertWindsurfRequestToClaude,
		interfaces.TranslateResponse{
			Stream:     ConvertClaudeResponseToWindsurf,
			NonStream:  ConvertClaudeResponseToWindsurfNonStream,
			TokenCount: WindsurfTokenCount,
		},
	)
}

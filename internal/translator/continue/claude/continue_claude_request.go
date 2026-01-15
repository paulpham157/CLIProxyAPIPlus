package claude

import (
	"bytes"
)

func ConvertContinueRequestToClaude(modelName string, inputRawJSON []byte, stream bool) []byte {
	rawJSON := bytes.Clone(inputRawJSON)
	return rawJSON
}

package continuetr

import (
	"bytes"
)

func ConvertClaudeRequestToContinue(modelName string, inputRawJSON []byte, stream bool) []byte {
	rawJSON := bytes.Clone(inputRawJSON)
	return rawJSON
}

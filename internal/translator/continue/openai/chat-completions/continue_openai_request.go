package chatcompletions

import (
	"bytes"
)

func ConvertContinueRequestToOpenAI(modelName string, inputRawJSON []byte, stream bool) []byte {
	rawJSON := bytes.Clone(inputRawJSON)
	return rawJSON
}

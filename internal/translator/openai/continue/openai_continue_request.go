package continuetr

import (
	"bytes"
)

func ConvertOpenAIRequestToContinue(modelName string, inputRawJSON []byte, stream bool) []byte {
	rawJSON := bytes.Clone(inputRawJSON)
	return rawJSON
}

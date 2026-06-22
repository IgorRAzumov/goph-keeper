package record

import (
	"encoding/json"
	"os"

	"goph-keeper/internal/client/model"
)

func buildPayload(input AddInput) ([]byte, error) {
	switch input.Type {
	case model.RecordTypeText:
		text, _ := input.Payload.(string)
		return json.Marshal(map[string]string{"text": text})
	case model.RecordTypeLogin:
		login, _ := input.Payload.(map[string]string)
		return json.Marshal(login)
	case model.RecordTypeCard:
		card, _ := input.Payload.(map[string]string)
		return json.Marshal(card)
	case model.RecordTypeBinary:
		if input.FilePath != "" {
			return os.ReadFile(input.FilePath)
		}
		if bytes, ok := input.Payload.([]byte); ok {
			return bytes, nil
		}
		return nil, model.ErrInvalidInput
	default:
		return json.Marshal(input.Payload)
	}
}

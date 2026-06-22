package record

import "goph-keeper/internal/client/model"

// AddInput — параметры создания записи.
type AddInput struct {
	Type     model.RecordType
	Meta     string
	Payload  any
	FilePath string
}

// View — расшифрованная запись для отображения в CLI.
type View struct {
	ID      string
	Type    model.RecordType
	Meta    string
	Payload string
}

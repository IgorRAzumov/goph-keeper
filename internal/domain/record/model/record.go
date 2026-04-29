package model

// RecordType классифицирует зашифрованные записи хранилища (логин, текст, бинарные данные, карта, опционально OTP).
type RecordType string

const (
	RecordTypeLogin  RecordType = "login"
	RecordTypeText   RecordType = "text"
	RecordTypeBinary RecordType = "binary"
	RecordTypeCard   RecordType = "card"
	RecordTypeOTP    RecordType = "otp"
)

// Record — унифицированная зашифрованная запись, хранимая для владельца (E2E ciphertext в API/БД).
type Record struct {
	// ID — идентификатор записи, генерируемый клиентом (непрозрачная строка).
	ID string
	// OwnerID — идентификатор владельца записи (пользователя).
	OwnerID string
	// Type определяет семантику payload после расшифровки на клиенте.
	Type RecordType
	// Meta — произвольная текстовая метаинформация (JSON или plain text).
	Meta string
	// Ciphertext — зашифрованные байты payload (для сервера непрозрачно).
	Ciphertext []byte
	// Version поддерживает optimistic concurrency между несколькими клиентами.
	Version int64
	// Deleted = true, когда запись является tombstone для синхронизации.
	Deleted bool
}

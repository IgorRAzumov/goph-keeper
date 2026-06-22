// Package model описывает доменные типы клиента GophKeeper.
//
// Клиент — самостоятельный bounded context: он не зависит от доменной модели
// сервера. Сетевой контракт (contract.Record) маппится в эти типы на границе
// пакета client/app, поэтому изменения серверного домена не протекают в клиент.
package model

// RecordType классифицирует приватные записи пользователя на стороне клиента.
type RecordType string

// Поддерживаемые типы записей.
const (
	// RecordTypeLogin — пара логин/пароль.
	RecordTypeLogin RecordType = "login"
	// RecordTypeText — произвольный текст.
	RecordTypeText RecordType = "text"
	// RecordTypeBinary — произвольные бинарные данные.
	RecordTypeBinary RecordType = "binary"
	// RecordTypeCard — данные банковской карты.
	RecordTypeCard RecordType = "card"
	// RecordTypeOTP — секрет одноразовых паролей (one-time password).
	RecordTypeOTP RecordType = "otp"
)

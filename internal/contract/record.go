// Package contract описывает JSON-контракт HTTP API GophKeeper.
//
// Это единственный источник истины для формата auth и sync на проводе: на него
// ссылаются и серверный слой доставки, и клиентский API-клиент, чтобы поля
// не расходились между сторонами.
package contract

// Record — зашифрованная запись в теле sync API. Ciphertext — base64 (std-кодировка).
type Record struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Meta       string `json:"meta"`
	Ciphertext string `json:"ciphertext"`
	Version    int64  `json:"version"`
	Deleted    bool   `json:"deleted"`
}

// PushRequest — тело POST /api/v1/sync (изменения клиента на сервер).
type PushRequest struct {
	Records []Record `json:"records"`
}

// PullResponse — тело ответа GET /api/v1/sync (изменения сервера клиенту).
type PullResponse struct {
	Records []Record `json:"records"`
}

// PushResponse — тело ответа POST /api/v1/sync (конфликтующие версии).
type PushResponse struct {
	Conflicts []Record `json:"conflicts"`
}

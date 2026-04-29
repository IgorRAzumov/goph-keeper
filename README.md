## GophKeeper

GophKeeper — клиент-серверный менеджер приватных данных (CLI + API сервер).

### Статус

Пока реализован каркас API сервера (`cmd/server` + composition root `internal/app`): `GET /healthz` и заглушки `/api/v1/auth/*`, `/api/v1/sync`. Структура каталогов следует чистой архитектуре: `domain`, `application`, `infrastructure`, `delivery/http`, `config`.

### Запуск сервера (dev)

```bash
go run ./cmd/server
```

Проверка:

```bash
curl -i http://127.0.0.1:8080/healthz
```


## GophKeeper

GophKeeper — клиент-серверный менеджер приватных данных (CLI + API сервер).

### Запуск сервера (dev)

По умолчанию сервер подключается к:

`postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable`

```bash
go run ./cmd/server
```

Опционально — переменные окружения:

```bash
export GOPHKEEPER_POSTGRES_DSN="postgres://postgres:ВАШ_ПАРОЛЬ@127.0.0.1:5432/postgres?sslmode=disable"
export GOPHKEEPER_JWT_SECRET="local-dev-jwt-secret"
go run ./cmd/server
```

Проверка:

```bash
curl -i http://127.0.0.1:8080/healthz
```


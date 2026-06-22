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

### CLI-клиент

Сборка:

```bash
go build -ldflags "-X goph-keeper/internal/buildinfo.Version=0.1.0 -X goph-keeper/internal/buildinfo.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o gophkeeper ./cmd/client
```

Команды:

```bash
gophkeeper version
gophkeeper register -login alice -password secret
gophkeeper login -login alice
gophkeeper sync
gophkeeper add -type text -meta "note" -data "hello"
gophkeeper list
gophkeeper get <id>
gophkeeper delete <id>
gophkeeper logout
```

Переменные окружения клиента:

| Переменная                   | Описание                                       |
|------------------------------|------------------------------------------------|
| `GOPHKEEPER_SERVER_URL`      | URL API (по умолчанию `http://127.0.0.1:8080`) |
| `GOPHKEEPER_MASTER_PASSWORD` | мастер-пароль для E2E-шифрования               |
| `GOPHKEEPER_CONFIG_DIR`      | каталог конфига (по умолчанию `~/.gophkeeper`) |

Конфиг и локальные данные: `~/.gophkeeper/config.json`, `~/.gophkeeper/data.json`.

### API (кратко)

| Метод  | Путь                    | Описание                  |
|--------|-------------------------|---------------------------|
| `POST` | `/api/v1/auth/register` | Регистрация               |
| `POST` | `/api/v1/auth/login`    | Логин → access + refresh  |
| `POST` | `/api/v1/auth/refresh`  | Обновление токенов        |
| `POST` | `/api/v1/auth/logout`   | Выход                     |
| `GET`  | `/api/v1/sync/?since=0` | Выгрузка записей (Bearer) |
| `POST` | `/api/v1/sync/`         | Загрузка записей (Bearer) |


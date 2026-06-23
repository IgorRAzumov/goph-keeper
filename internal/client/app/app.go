package app

import (
	"context"

	appauth "goph-keeper/internal/client/application/auth"
	apprecord "goph-keeper/internal/client/application/record"
	"goph-keeper/internal/client/application/session"
	"goph-keeper/internal/client/application/state"
	appsync "goph-keeper/internal/client/application/sync"
)

// AddInput — параметры новой записи (алиас application/record).
type AddInput = apprecord.AddInput

// DisplayRecord — запись для вывода в CLI (алиас application/record).
type DisplayRecord = apprecord.View

// App — composition root клиента: делегирует сценарии в application layer.
type App struct {
	*state.State
	auth   *appauth.Usecase
	sync   *appsync.Usecase
	record *apprecord.Usecase
}

// New загружает конфиг, хранилище и собирает usecase'ы.
func New(masterPassword, serverURL string) (*App, error) {
	state, err := state.Load(masterPassword, serverURL)
	if err != nil {
		return nil, err
	}
	session := session.New(state)
	syncUC := appsync.New(state, session)
	return &App{
		State:  state,
		auth:   appauth.New(state, session),
		sync:   syncUC,
		record: apprecord.New(state, session, syncUC),
	}, nil
}

// Save сохраняет конфиг и данные.
func (app *App) Save() error {
	return app.State.Save()
}

// Register регистрирует пользователя на сервере.
func (app *App) Register(ctx context.Context, login, password string) error {
	return app.auth.Register(ctx, login, password)
}

// Login аутентифицирует пользователя.
func (app *App) Login(ctx context.Context, login, password string) error {
	return app.auth.Login(ctx, login, password)
}

// Logout завершает сессию.
func (app *App) Logout(ctx context.Context) error {
	return app.auth.Logout(ctx)
}

// Sync выполняет pull и push.
func (app *App) Sync(ctx context.Context) error {
	return app.sync.Run(ctx)
}

// Add создаёт запись локально.
func (app *App) Add(ctx context.Context, in AddInput) (string, error) {
	return app.record.Add(ctx, in)
}

// List возвращает расшифрованные активные записи.
func (app *App) List(ctx context.Context) ([]DisplayRecord, error) {
	return app.record.List(ctx)
}

// Get возвращает одну расшифрованную запись.
func (app *App) Get(ctx context.Context, id string) (DisplayRecord, error) {
	return app.record.Get(ctx, id)
}

// Delete помечает запись удалённой и синхронизирует.
func (app *App) Delete(ctx context.Context, id string) error {
	return app.record.Delete(ctx, id)
}

package record

import (
	"context"
	"errors"
	"fmt"

	"goph-keeper/internal/client/api"
	appsecrets "goph-keeper/internal/client/application/secrets"
	"goph-keeper/internal/client/application/session"
	"goph-keeper/internal/client/application/state"
	appsync "goph-keeper/internal/client/application/sync"
	clientcrypto "goph-keeper/internal/client/crypto"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/client/store"
)

// Usecase управляет локальными записями и их синхронизацией.
type Usecase struct {
	State   *state.State
	Session *session.Service
	Sync    *appsync.Usecase
}

// New создаёт record usecase.
func New(state *state.State, session *session.Service, sync *appsync.Usecase) *Usecase {
	return &Usecase{State: state, Session: session, Sync: sync}
}

// Add создаёт запись локально и синхронизирует с сервером.
func (usecase *Usecase) Add(ctx context.Context, in AddInput) (string, error) {
	if err := usecase.Session.Ensure(ctx); err != nil {
		return "", err
	}
	key, err := appsecrets.MasterKey(usecase.State)
	if err != nil {
		return "", err
	}
	plain, err := buildPayload(in)
	if err != nil {
		return "", err
	}
	ciphertext, err := clientcrypto.Encrypt(key, plain)
	if err != nil {
		return "", err
	}
	id, err := model.NewID()
	if err != nil {
		return "", err
	}
	record := store.Record{
		ID:         id,
		Type:       in.Type,
		Meta:       in.Meta,
		Ciphertext: ciphertext,
		Version:    usecase.State.Data.NextVersion(),
	}
	usecase.State.Data.Put(record)
	if err := usecase.State.Save(); err != nil {
		return "", err
	}
	if err := usecase.Sync.Run(ctx); err != nil && !errors.Is(err, api.ErrConflict) {
		return id, err
	}
	return id, nil
}

// List возвращает расшифрованные активные записи из локального хранилища.
func (usecase *Usecase) List(ctx context.Context) ([]View, error) {
	if err := usecase.Session.Ensure(ctx); err != nil {
		return nil, err
	}
	key, err := appsecrets.MasterKey(usecase.State)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0)
	for _, record := range usecase.State.Data.ListActive() {
		plain, err := clientcrypto.Decrypt(key, record.Ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decrypt %s: %w", record.ID, err)
		}
		out = append(out, View{
			ID: record.ID, Type: record.Type, Meta: record.Meta, Payload: string(plain),
		})
	}
	return out, nil
}

// Get возвращает одну расшифрованную запись по id из локального хранилища.
func (usecase *Usecase) Get(ctx context.Context, id string) (View, error) {
	if err := usecase.Session.Ensure(ctx); err != nil {
		return View{}, err
	}
	record, err := usecase.State.Data.Get(id)
	if err != nil {
		return View{}, err
	}
	key, err := appsecrets.MasterKey(usecase.State)
	if err != nil {
		return View{}, err
	}
	plain, err := clientcrypto.Decrypt(key, record.Ciphertext)
	if err != nil {
		return View{}, err
	}
	return View{ID: record.ID, Type: record.Type, Meta: record.Meta, Payload: string(plain)}, nil
}

// Delete помечает запись удалённой и синхронизирует.
func (usecase *Usecase) Delete(ctx context.Context, id string) error {
	if err := usecase.Session.Ensure(ctx); err != nil {
		return err
	}
	if err := usecase.State.Data.Delete(id); err != nil {
		return err
	}
	if err := usecase.State.Save(); err != nil {
		return err
	}
	return usecase.Sync.Run(ctx)
}

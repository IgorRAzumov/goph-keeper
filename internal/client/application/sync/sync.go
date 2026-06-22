package sync

import (
	"context"
	"errors"

	"goph-keeper/internal/client/api"
	"goph-keeper/internal/client/application/session"
	"goph-keeper/internal/client/application/state"
	"goph-keeper/internal/contract"
)

// Usecase выполняет pull/push синхронизации с сервером.
type Usecase struct {
	State   *state.State
	Session *session.Service
}

// New создаёт sync usecase.
func New(state *state.State, session *session.Service) *Usecase {
	return &Usecase{State: state, Session: session}
}

// Run выполняет pull и push локальных изменений.
func (usecase *Usecase) Run(ctx context.Context) error {
	if err := usecase.Session.Ensure(ctx); err != nil {
		return err
	}

	pulled, err := usecase.pullWithRetry(ctx)
	if err != nil {
		return err
	}
	for _, wire := range pulled {
		record, err := WireToRecord(wire)
		if err != nil {
			return err
		}
		usecase.State.Data.MergeRemote(record)
	}

	if err := usecase.pushDirtyWithRebase(ctx); err != nil {
		return err
	}
	return usecase.State.Save()
}

func (usecase *Usecase) pullWithRetry(ctx context.Context) ([]contract.Record, error) {
	pulled, err := usecase.State.API.Pull(ctx, usecase.State.Config.AccessToken, usecase.State.Data.LastSyncVersion)
	if err == nil {
		return pulled, nil
	}
	if refreshErr := usecase.Session.Refresh(ctx); refreshErr != nil {
		return nil, err
	}
	return usecase.State.API.Pull(ctx, usecase.State.Config.AccessToken, usecase.State.Data.LastSyncVersion)
}

func (usecase *Usecase) pushDirtyWithRebase(ctx context.Context) error {
	for attempt := 0; attempt < 2; attempt++ {
		dirty := usecase.State.Data.DirtyRecords()
		if len(dirty) == 0 {
			return nil
		}

		wires := make([]contract.Record, 0, len(dirty))
		ids := make([]string, 0, len(dirty))
		for _, record := range dirty {
			wires = append(wires, RecordToWire(record))
			ids = append(ids, record.ID)
		}

		conflicts, err := usecase.pushWithRetry(ctx, wires)
		if err == nil {
			usecase.State.Data.MarkPushed(ids...)
			return nil
		}
		if !errors.Is(err, api.ErrConflict) {
			return err
		}

		for _, wire := range conflicts {
			record, convErr := WireToRecord(wire)
			if convErr != nil {
				return convErr
			}
			usecase.State.Data.MergeRemote(record)
		}

		if len(conflicts) == 0 {
			rebased := usecase.State.Data.RebaseDirtyVersions()
			if len(rebased) == 0 {
				return err
			}
		}
	}
	return api.ErrConflict
}

func (usecase *Usecase) pushWithRetry(ctx context.Context, records []contract.Record) ([]contract.Record, error) {
	conflicts, err := usecase.State.API.Push(ctx, usecase.State.Config.AccessToken, records)
	if err == nil || errors.Is(err, api.ErrConflict) {
		return conflicts, err
	}
	if refreshErr := usecase.Session.Refresh(ctx); refreshErr != nil {
		return nil, err
	}
	return usecase.State.API.Push(ctx, usecase.State.Config.AccessToken, records)
}

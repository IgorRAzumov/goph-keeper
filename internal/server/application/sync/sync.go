package sync

import (
	"context"
	"strings"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/record/model"
	recordsvc "goph-keeper/internal/server/domain/record/service"
)

// Usecase выполняет pull/push синхронизации зашифрованных записей владельца.
type Usecase struct {
	recordService *recordsvc.RecordService
}

// NewUsecase создаёт сценарий синхронизации.
func NewUsecase(recordService *recordsvc.RecordService) *Usecase {
	return &Usecase{recordService: recordService}
}

// PullInput — параметры выгрузки изменений с сервера.
type PullInput struct {
	OwnerID      string
	SinceVersion int64
}

// PullOutput — записи владельца, изменённые после SinceVersion.
type PullOutput struct {
	Records []*model.Record
}

// Pull возвращает записи с version > SinceVersion.
func (usecase *Usecase) Pull(ctx context.Context, input PullInput) (PullOutput, error) {
	if strings.TrimSpace(input.OwnerID) == "" || input.SinceVersion < 0 {
		return PullOutput{}, common.ErrInvalidInput
	}
	if usecase == nil || usecase.recordService == nil {
		return PullOutput{}, common.ErrNotImplemented
	}

	records, err := usecase.recordService.ListSince(ctx, input.OwnerID, input.SinceVersion)
	if err != nil {
		return PullOutput{}, err
	}
	return PullOutput{Records: records}, nil
}

// PushInput — набор записей, которые клиент отправляет на сервер.
type PushInput struct {
	OwnerID string
	Records []*model.Record
}

// PushOutput — конфликты: на сервере уже есть более новая версия.
type PushOutput struct {
	Conflicts []*model.Record
}

// Push применяет изменения клиента атомарно (last-write-wins по version).
func (usecase *Usecase) Push(ctx context.Context, input PushInput) (PushOutput, error) {
	if strings.TrimSpace(input.OwnerID) == "" {
		return PushOutput{}, common.ErrInvalidInput
	}
	if usecase == nil || usecase.recordService == nil {
		return PushOutput{}, common.ErrNotImplemented
	}

	conflicts, err := usecase.recordService.UpdateBatch(ctx, input.OwnerID, input.Records)
	if err != nil {
		return PushOutput{}, err
	}
	if conflicts == nil {
		conflicts = []*model.Record{}
	}
	return PushOutput{Conflicts: conflicts}, nil
}

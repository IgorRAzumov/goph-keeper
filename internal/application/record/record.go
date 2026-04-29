package record

import (
	"context"
	"strings"

	"goph-keeper/internal/domain/common"
	recordsvc "goph-keeper/internal/domain/record/service"
)

// Usecase объединяет сценарии работы с приватными записями.
type Usecase struct {
	recordService *recordsvc.RecordService
}

// NewRecordUseCase создаёт сценарий работы с приватными записями.
func NewRecordUseCase(recordService *recordsvc.RecordService) *Usecase {
	return &Usecase{recordService: recordService}
}

// Read возвращает запись по идентификатору.
func (usecase *Usecase) Read(ctx context.Context, input GetRecordInput) (GetRecordOutput, error) {
	if strings.TrimSpace(input.OwnerID) == "" || strings.TrimSpace(input.RecordID) == "" {
		return GetRecordOutput{}, common.ErrInvalidInput
	}
	if usecase == nil || usecase.recordService == nil {
		return GetRecordOutput{}, common.ErrNotImplemented
	}

	record, err := usecase.recordService.Get(ctx, input.OwnerID, input.RecordID)
	if err != nil {
		return GetRecordOutput{}, err
	}
	return GetRecordOutput{Record: record}, nil
}

// Update создаёт или обновляет запись.
func (usecase *Usecase) Update(ctx context.Context, input UpdateRecordInput) (UpdateRecordOutput, error) {
	if strings.TrimSpace(input.OwnerID) == "" || input.Record == nil || strings.TrimSpace(input.Record.ID) == "" {
		return UpdateRecordOutput{}, common.ErrInvalidInput
	}
	if usecase == nil || usecase.recordService == nil {
		return UpdateRecordOutput{}, common.ErrNotImplemented
	}

	if err := usecase.recordService.Update(ctx, input.OwnerID, input.Record); err != nil {
		return UpdateRecordOutput{}, err
	}
	return UpdateRecordOutput{}, nil
}

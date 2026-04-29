package record

import (
	"context"
	"strings"

	"goph-keeper/internal/domain/common"
	recordsvc "goph-keeper/internal/domain/record/service"
)

// UpdateRecordUseCase создаёт или обновляет запись (сценарий записи).
type UpdateRecordUseCase struct {
	recordService *recordsvc.RecordService
}

// NewUpdateRecordUseCase создаёт сценарий обновления записи.
func NewUpdateRecordUseCase(recordService *recordsvc.RecordService) *UpdateRecordUseCase {
	return &UpdateRecordUseCase{recordService: recordService}
}

func (usecase *UpdateRecordUseCase) Execute(ctx context.Context, in UpdateRecordInput) (UpdateRecordOutput, error) {
	if strings.TrimSpace(in.OwnerID) == "" || in.Record == nil || strings.TrimSpace(in.Record.ID) == "" {
		return UpdateRecordOutput{}, common.ErrInvalidInput
	}
	if usecase == nil || usecase.recordService == nil {
		return UpdateRecordOutput{}, common.ErrNotImplemented
	}

	if err := usecase.recordService.Update(ctx, in.OwnerID, in.Record); err != nil {
		return UpdateRecordOutput{}, err
	}
	return UpdateRecordOutput{}, nil
}

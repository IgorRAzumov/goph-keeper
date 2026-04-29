package record

import (
	"context"
	"strings"

	"goph-keeper/internal/domain/common"
	recordsvc "goph-keeper/internal/domain/record/service"
)

// GetRecordUseCase возвращает запись по идентификатору (сценарий чтения).
type GetRecordUseCase struct {
	recordService *recordsvc.RecordService
}

// NewGetRecordUseCase создаёт сценарий чтения записи.
func NewGetRecordUseCase(recordService *recordsvc.RecordService) *GetRecordUseCase {
	return &GetRecordUseCase{recordService: recordService}
}

func (usecase *GetRecordUseCase) Execute(ctx context.Context, input GetRecordInput) (GetRecordOutput, error) {
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

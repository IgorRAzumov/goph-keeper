package service

import (
	"context"
	"strings"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/record/model"
	recordrepo "goph-keeper/internal/server/domain/record/repository"
)

// RecordService инкапсулирует доменные операции над зашифрованными записями.
// Сценарии application-слоя вызывают сервис, а сервис работает через порт репозитория.
type RecordService struct {
	recordRepository recordrepo.RecordRepository
}

// NewRecordService создаёт сервис записей.
func NewRecordService(recordRepository recordrepo.RecordRepository) *RecordService {
	return &RecordService{recordRepository: recordRepository}
}

// Get возвращает запись по идентификатору для владельца.
func (service *RecordService) Get(ctx context.Context, ownerID, recordID string) (*model.Record, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(recordID) == "" {
		return nil, common.ErrInvalidInput
	}
	if service == nil || service.recordRepository == nil {
		return nil, common.ErrNotImplemented
	}
	return service.recordRepository.Get(ctx, ownerID, recordID)
}

// ListSince возвращает записи владельца, изменённые после sinceVersion.
func (service *RecordService) ListSince(ctx context.Context, ownerID string, sinceVersion int64) ([]*model.Record, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, common.ErrInvalidInput
	}
	if sinceVersion < 0 {
		return nil, common.ErrInvalidInput
	}
	if service == nil || service.recordRepository == nil {
		return nil, common.ErrNotImplemented
	}
	return service.recordRepository.ListSince(ctx, ownerID, sinceVersion)
}

// Update создаёт или обновляет запись для владельца.
func (service *RecordService) Update(ctx context.Context, ownerID string, record *model.Record) error {
	if strings.TrimSpace(ownerID) == "" || record == nil || strings.TrimSpace(record.ID) == "" {
		return common.ErrInvalidInput
	}
	ownerID = strings.TrimSpace(ownerID)
	if strings.TrimSpace(record.OwnerID) != "" && record.OwnerID != ownerID {
		return common.ErrInvalidInput
	}
	if !validRecordType(record.Type) || (!record.Deleted && len(record.Ciphertext) == 0) {
		return common.ErrInvalidInput
	}
	if service == nil || service.recordRepository == nil {
		return common.ErrNotImplemented
	}
	record.OwnerID = ownerID
	return service.recordRepository.Update(ctx, ownerID, record)
}

// UpdateBatch применяет набор записей в одной транзакции и возвращает конфликтующие версии.
func (service *RecordService) UpdateBatch(ctx context.Context, ownerID string, records []*model.Record) ([]*model.Record, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, common.ErrInvalidInput
	}
	if service == nil || service.recordRepository == nil {
		return nil, common.ErrNotImplemented
	}
	ownerID = strings.TrimSpace(ownerID)

	normalized := make([]*model.Record, 0, len(records))
	for _, record := range records {
		if record == nil || strings.TrimSpace(record.ID) == "" {
			return nil, common.ErrInvalidInput
		}
		if strings.TrimSpace(record.OwnerID) != "" && record.OwnerID != ownerID {
			return nil, common.ErrInvalidInput
		}
		if !validRecordType(record.Type) || (!record.Deleted && len(record.Ciphertext) == 0) {
			return nil, common.ErrInvalidInput
		}
		record.OwnerID = ownerID
		normalized = append(normalized, record)
	}
	return service.recordRepository.UpdateBatch(ctx, ownerID, normalized)
}

func validRecordType(recordType model.RecordType) bool {
	switch recordType {
	case
		model.RecordTypeLogin,
		model.RecordTypeText,
		model.RecordTypeBinary,
		model.RecordTypeCard,
		model.RecordTypeOTP:
		return true
	default:
		return false
	}
}

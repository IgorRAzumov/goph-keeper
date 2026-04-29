package service

import (
	"context"
	"strings"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/record/model"
	recordrepo "goph-keeper/internal/domain/record/repository"
)

// RecordService инкапсулирует доменные операции над зашифрованными записями.
// Сценарии application-слоя вызывают сервис, а сервис работает через порт репозитория.
type RecordService struct {
	repo recordrepo.RecordRepository
}

// NewRecordService создаёт сервис записей.
func NewRecordService(repo recordrepo.RecordRepository) *RecordService {
	return &RecordService{repo: repo}
}

// Get возвращает запись по идентификатору для владельца.
func (s *RecordService) Get(ctx context.Context, ownerID, recordID string) (*model.Record, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(recordID) == "" {
		return nil, common.ErrInvalidInput
	}
	if s == nil || s.repo == nil {
		return nil, common.ErrNotImplemented
	}
	return s.repo.Get(ctx, ownerID, recordID)
}

// Update создаёт или обновляет запись для владельца.
func (s *RecordService) Update(ctx context.Context, ownerID string, r *model.Record) error {
	if strings.TrimSpace(ownerID) == "" || r == nil || strings.TrimSpace(r.ID) == "" {
		return common.ErrInvalidInput
	}
	ownerID = strings.TrimSpace(ownerID)
	if strings.TrimSpace(r.OwnerID) != "" && r.OwnerID != ownerID {
		return common.ErrInvalidInput
	}
	if !validRecordType(r.Type) || (!r.Deleted && len(r.Ciphertext) == 0) {
		return common.ErrInvalidInput
	}
	if s == nil || s.repo == nil {
		return common.ErrNotImplemented
	}
	r.OwnerID = ownerID
	return s.repo.Update(ctx, ownerID, r)
}

func validRecordType(recordType model.RecordType) bool {
	switch recordType {
	case model.RecordTypeLogin, model.RecordTypeText, model.RecordTypeBinary, model.RecordTypeCard, model.RecordTypeOTP:
		return true
	default:
		return false
	}
}

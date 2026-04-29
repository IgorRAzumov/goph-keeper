package repository

import (
	"context"

	"goph-keeper/internal/domain/record/model"
)

// RecordRepository хранит зашифрованные записи и поддерживает операции синхронизации (порт).
type RecordRepository interface {
	// Update создаёт или обновляет запись для указанного владельца.
	Update(ctx context.Context, ownerID string, r *model.Record) error
	// Get возвращает запись по id для владельца или ErrNotFound.
	Get(ctx context.Context, ownerID, recordID string) (*model.Record, error)
}

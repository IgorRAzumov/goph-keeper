package repository

import (
	"context"

	"goph-keeper/internal/server/domain/record/model"
)

// RecordStore хранит зашифрованные записи и поддерживает операции синхронизации (порт).
type RecordStore interface {
	// Update создаёт или обновляет запись для указанного владельца (last-write-wins по version).
	Update(ctx context.Context, ownerID string, r *model.Record) error
	// UpdateBatch применяет набор записей в одной транзакции; возвращает конфликтующие версии (без partial commit при ошибке).
	UpdateBatch(ctx context.Context, ownerID string, records []*model.Record) ([]*model.Record, error)
	// Get возвращает запись по id для владельца или ErrNotFound.
	Get(ctx context.Context, ownerID, recordID string) (*model.Record, error)
	// ListSince возвращает записи владельца с version строго больше sinceVersion (глобальный watermark владельца).
	ListSince(ctx context.Context, ownerID string, sinceVersion int64) ([]*model.Record, error)
}

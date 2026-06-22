package repository

import (
	"context"
	"time"

	"goph-keeper/internal/server/domain/session/model"
)

// SessionRepository — порт хранения сессий.
type SessionRepository interface {
	// Save создаёт или обновляет сессию.
	Save(ctx context.Context, session *model.Session) error
	// Get возвращает сессию по id или ErrNotFound.
	Get(ctx context.Context, id string) (*model.Session, error)
	// Delete удаляет сессию (используется при ротации refresh).
	Delete(ctx context.Context, id string) error
	// Rotate атомарно удаляет старую сессию и создаёт новую.
	Rotate(ctx context.Context, oldSessionID string, newSession *model.Session) error
	// FindActiveByRefreshToken находит активную сессию по refresh-токену (plain) и времени now.
	FindActiveByRefreshToken(ctx context.Context, refreshToken string, now time.Time) (*model.Session, error)
}

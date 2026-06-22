package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/session/model"
)

// SessionRepository — Postgres-адаптер для sessionrepository.SessionRepository.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository создаёт репозиторий сессий.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (repository *SessionRepository) Save(ctx context.Context, session *model.Session) error {
	if repository == nil || repository.db == nil {
		return common.ErrNotImplemented
	}
	if session == nil || session.ID == "" || session.UserID == "" || len(session.RefreshTokenHash) == 0 {
		return common.ErrInvalidInput
	}

	_, err := repository.db.ExecContext(ctx, `
INSERT INTO keeper_sessions (id, user_id, refresh_token_hash, refresh_expires_at)
VALUES ($1, $2, $3, $4)
`, session.ID, session.UserID, session.RefreshTokenHash, session.RefreshExpiresAt)
	return err
}

func (repository *SessionRepository) Get(ctx context.Context, id string) (*model.Session, error) {
	if repository == nil || repository.db == nil {
		return nil, common.ErrNotImplemented
	}
	if id == "" {
		return nil, common.ErrInvalidInput
	}

	row := repository.db.QueryRowContext(ctx, `
SELECT id, user_id, refresh_token_hash, refresh_expires_at
FROM keeper_sessions
WHERE id=$1
LIMIT 1
`, id)

	var session model.Session
	if err := row.Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.RefreshExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (repository *SessionRepository) Delete(ctx context.Context, id string) error {
	if repository == nil || repository.db == nil {
		return common.ErrNotImplemented
	}
	if id == "" {
		return common.ErrInvalidInput
	}
	_, err := repository.db.ExecContext(ctx, `DELETE FROM keeper_sessions WHERE id=$1`, id)
	return err
}

func (repository *SessionRepository) Rotate(ctx context.Context, oldSessionID string, newSession *model.Session) error {
	if repository == nil || repository.db == nil {
		return common.ErrNotImplemented
	}
	if oldSessionID == "" || newSession == nil || newSession.ID == "" ||
		newSession.UserID == "" || len(newSession.RefreshTokenHash) == 0 {
		return common.ErrInvalidInput
	}

	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()

	result, err := transaction.ExecContext(ctx, `DELETE FROM keeper_sessions WHERE id=$1`, oldSessionID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return common.ErrNotFound
	}

	_, err = transaction.ExecContext(ctx, `
INSERT INTO keeper_sessions (id, user_id, refresh_token_hash, refresh_expires_at)
VALUES ($1, $2, $3, $4)
`, newSession.ID, newSession.UserID, newSession.RefreshTokenHash, newSession.RefreshExpiresAt)
	if err != nil {
		return err
	}
	return transaction.Commit()
}

func (repository *SessionRepository) FindActiveByRefreshToken(ctx context.Context, refreshToken string, now time.Time) (*model.Session, error) {
	if repository == nil || repository.db == nil {
		return nil, common.ErrNotImplemented
	}
	if refreshToken == "" {
		return nil, common.ErrInvalidInput
	}

	row := repository.db.QueryRowContext(ctx, `
SELECT id, user_id, refresh_token_hash, refresh_expires_at
FROM keeper_sessions
WHERE refresh_token_hash=$1 AND refresh_expires_at > $2
LIMIT 1
`, model.HashRefreshToken(refreshToken), now)

	var session model.Session
	if err := row.Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.RefreshExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

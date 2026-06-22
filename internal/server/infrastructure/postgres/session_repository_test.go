package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"goph-keeper/internal/server/domain/common"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
)

func TestSessionRepositoryFindActiveByRefreshTokenUsesHashLookup(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewSessionRepository(db)
	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{"id", "user_id", "refresh_token_hash", "refresh_expires_at"}).
		AddRow("session-1", "user-1", sessionmodel.HashRefreshToken("refresh-token"), now.Add(time.Hour))

	mock.ExpectQuery(`WHERE refresh_token_hash=\$1 AND refresh_expires_at > \$2`).
		WithArgs(sessionmodel.HashRefreshToken("refresh-token"), now).
		WillReturnRows(rows)

	session, err := repo.FindActiveByRefreshToken(context.Background(), "refresh-token", now)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if session.ID != "session-1" || session.UserID != "user-1" {
		t.Fatalf("unexpected session: %#v", session)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionRepositoryRotateRunsInTransaction(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewSessionRepository(db)
	newSession := &sessionmodel.Session{
		ID:               "new-session",
		UserID:           "user-1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("new-refresh"),
		RefreshExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM keeper_sessions WHERE id=\$1`).
		WithArgs("old-session").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO keeper_sessions`).
		WithArgs(newSession.ID, newSession.UserID, newSession.RefreshTokenHash, newSession.RefreshExpiresAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.Rotate(context.Background(), "old-session", newSession); err != nil {
		t.Fatalf("rotate failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionRepositoryRotateReturnsNotFoundWhenOldSessionMissing(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewSessionRepository(db)
	newSession := &sessionmodel.Session{
		ID:               "new-session",
		UserID:           "user-1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("new-refresh"),
		RefreshExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM keeper_sessions WHERE id=\$1`).
		WithArgs("missing-session").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	if err := repo.Rotate(context.Background(), "missing-session", newSession); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

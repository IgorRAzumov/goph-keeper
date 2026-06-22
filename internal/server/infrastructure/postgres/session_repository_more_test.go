package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"goph-keeper/internal/server/domain/common"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
)

func TestSessionRepositorySaveAndGet(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewSessionRepository(db)
	exp := time.Now().UTC().Add(time.Hour)
	sess := &sessionmodel.Session{
		ID: "s1", UserID: "u1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("rt"),
		RefreshExpiresAt: exp,
	}

	mock.ExpectExec(`INSERT INTO keeper_sessions`).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.Save(context.Background(), sess); err != nil {
		t.Fatal(err)
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "refresh_token_hash", "refresh_expires_at"}).
		AddRow("s1", "u1", sess.RefreshTokenHash, exp)
	mock.ExpectQuery(`FROM keeper_sessions`).WithArgs("s1").WillReturnRows(rows)

	got, err := repo.Get(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != "u1" {
		t.Fatalf("unexpected session: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionRepositoryNilDB(t *testing.T) {
	t.Parallel()

	repo := NewSessionRepository(nil)
	if err := repo.Save(context.Background(), &sessionmodel.Session{ID: "s"}); !isNotImplemented(err) {
		t.Fatalf("expected not implemented, got %v", err)
	}
}

func isNotImplemented(err error) bool {
	return err == common.ErrNotImplemented
}

func TestSessionRepositoryDelete(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectExec(`DELETE FROM keeper_sessions`).WithArgs("s1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := NewSessionRepository(db).Delete(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	_ = sql.ErrNoRows
}

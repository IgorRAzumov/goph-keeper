package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"
)

func TestUserRepositoryNilDBReturnsNotImplemented(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository(nil)

	if err := repo.Save(context.Background(), &model.User{ID: "1", Login: "a", PasswordHash: []byte("x")}); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from Save, got %v", err)
	}
	if _, err := repo.GetByLogin(context.Background(), "alice"); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from GetByLogin, got %v", err)
	}
}

func TestUserRepositoryGetByLogin(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserRepository(db)
	rows := sqlmock.NewRows([]string{"id", "login", "password_hash"}).AddRow("u1", "alice", []byte("hash"))

	mock.ExpectQuery(`SELECT id, login, password_hash`).WithArgs("alice").WillReturnRows(rows)

	u, err := repo.GetByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != "u1" || u.Login != "alice" || string(u.PasswordHash) != "hash" {
		t.Fatalf("unexpected user: %+v", u)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepositorySaveConflictOnUniqueViolation(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT 1 FROM users WHERE login=`).WithArgs("dup").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO users`).WillReturnError(&pgconn.PgError{Code: "23505"})
	mock.ExpectRollback()

	err = repo.Save(context.Background(), &model.User{ID: "id1", Login: "dup", PasswordHash: []byte("h")})
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

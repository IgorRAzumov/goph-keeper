package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/user/model"
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
	rows := sqlmock.NewRows([]string{"id", "login", "password_hash", "master_salt"}).AddRow("u1", "alice", []byte("hash"), "salt-1")

	mock.ExpectQuery(`SELECT id, login, password_hash`).WithArgs("alice").WillReturnRows(rows)

	u, err := repo.GetByLogin(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != "u1" || u.Login != "alice" || string(u.PasswordHash) != "hash" || u.MasterSalt != "salt-1" {
		t.Fatalf("unexpected user: %+v", u)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepositoryGetByLoginNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT id, login, password_hash`).WithArgs("missing").WillReturnError(sql.ErrNoRows)
	_, err = NewUserRepository(db).GetByLogin(context.Background(), "missing")
	if !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
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

	mock.ExpectExec(`INSERT INTO keeper_users`).WillReturnError(&pgconn.PgError{Code: "23505"})

	err = repo.Save(context.Background(), &model.User{ID: "id1", Login: "dup", PasswordHash: []byte("h"), MasterSalt: "salt"})
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepositorySetMasterSalt(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserRepository(db)
	mock.ExpectExec(`UPDATE keeper_users`).WithArgs("u1", "salt").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.SetMasterSalt(context.Background(), "u1", "salt"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepositorySetMasterSaltNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewUserRepository(db)
	mock.ExpectExec(`UPDATE keeper_users`).WithArgs("u1", "salt").WillReturnResult(sqlmock.NewResult(0, 0))
	err = repo.SetMasterSalt(context.Background(), "u1", "salt")
	if !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/record/model"
)

func TestRecordRepositoryNilDBReturnsNotImplemented(t *testing.T) {
	t.Parallel()

	repo := NewRecordRepository(nil)
	record := &model.Record{ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("x"), Version: 1}

	if _, err := repo.Get(context.Background(), "owner", "r1"); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("Get: expected ErrNotImplemented, got %v", err)
	}
	if err := repo.Update(context.Background(), "owner", record); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("Update: expected ErrNotImplemented, got %v", err)
	}
	if _, err := repo.UpdateBatch(context.Background(), "owner", []*model.Record{record}); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("UpdateBatch: expected ErrNotImplemented, got %v", err)
	}
	if _, err := repo.ListSince(context.Background(), "owner", 0); !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("ListSince: expected ErrNotImplemented, got %v", err)
	}
}

func TestRecordRepositoryGet(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{"id", "owner_id", "type", "meta", "ciphertext", "version", "deleted"}).
		AddRow("r1", "owner-1", "text", "meta", []byte("cipher"), int64(3), false)
	mock.ExpectQuery(`SELECT id, owner_id, type, meta, ciphertext, version, deleted`).
		WithArgs("r1", "owner-1").
		WillReturnRows(rows)

	repo := NewRecordRepository(db)
	record, err := repo.Get(context.Background(), "owner-1", "r1")
	if err != nil {
		t.Fatal(err)
	}
	if record.ID != "r1" || record.OwnerID != "owner-1" || record.Version != 3 {
		t.Fatalf("unexpected record: %+v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryListSince(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	rows := sqlmock.NewRows([]string{"id", "owner_id", "type", "meta", "ciphertext", "version", "deleted"}).
		AddRow("r1", "owner-1", "text", "", []byte("c"), int64(2), false).
		AddRow("r2", "owner-1", "login", "", []byte("d"), int64(5), true)
	mock.ExpectQuery(`SELECT id, owner_id, type, meta, ciphertext, version, deleted`).
		WithArgs("owner-1", int64(1)).
		WillReturnRows(rows)

	repo := NewRecordRepository(db)
	records, err := repo.ListSince(context.Background(), "owner-1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryUpdateInsertsNewRecord(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT version FROM keeper_records`).
		WithArgs("r1", "owner-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version\), 0\) FROM keeper_records`).
		WithArgs("owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(int64(0)))
	mock.ExpectExec(`INSERT INTO keeper_records`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewRecordRepository(db)
	err = repo.Update(context.Background(), "owner-1", &model.Record{
		ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("c"), Version: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryInsertRejectsStaleGlobalVersion(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT version FROM keeper_records`).
		WithArgs("r2", "owner-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version\), 0\) FROM keeper_records`).
		WithArgs("owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(int64(10)))
	mock.ExpectRollback()

	repo := NewRecordRepository(db)
	err = repo.Update(context.Background(), "owner-1", &model.Record{
		ID: "r2", Type: model.RecordTypeText, Ciphertext: []byte("c"), Version: 5,
	})
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryUpdateRejectsStaleVersion(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT version FROM keeper_records`).
		WithArgs("r1", "owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(int64(5)))
	mock.ExpectQuery(`SELECT id, owner_id, type, meta, ciphertext, version, deleted`).
		WithArgs("r1", "owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id", "type", "meta", "ciphertext", "version", "deleted"}).
			AddRow("r1", "owner-1", "text", "", []byte("c"), int64(5), false))
	mock.ExpectRollback()

	repo := NewRecordRepository(db)
	err = repo.Update(context.Background(), "owner-1", &model.Record{
		ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("c"), Version: 5,
	})
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryUpdateAppliesNewerVersion(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT version FROM keeper_records`).
		WithArgs("r1", "owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(int64(5)))
	mock.ExpectExec(`UPDATE keeper_records`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := NewRecordRepository(db)
	err = repo.Update(context.Background(), "owner-1", &model.Record{
		ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("c"), Version: 6,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryUpdateDetectsConcurrentRace(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT version FROM keeper_records`).
		WithArgs("r1", "owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(int64(5)))
	mock.ExpectExec(`UPDATE keeper_records`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT id, owner_id, type, meta, ciphertext, version, deleted`).
		WithArgs("r1", "owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id", "type", "meta", "ciphertext", "version", "deleted"}).
			AddRow("r1", "owner-1", "text", "", []byte("c"), int64(7), false))
	mock.ExpectRollback()

	repo := NewRecordRepository(db)
	err = repo.Update(context.Background(), "owner-1", &model.Record{
		ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("c"), Version: 6,
	})
	if !errors.Is(err, common.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryUpdateBatchCollectsConflicts(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT version FROM keeper_records`).
		WithArgs("r1", "owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(int64(5)))
	mock.ExpectQuery(`SELECT id, owner_id, type, meta, ciphertext, version, deleted`).
		WithArgs("r1", "owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id", "type", "meta", "ciphertext", "version", "deleted"}).
			AddRow("r1", "owner-1", "text", "", []byte("server"), int64(5), false))
	mock.ExpectQuery(`SELECT version FROM keeper_records`).
		WithArgs("r2", "owner-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version\), 0\) FROM keeper_records`).
		WithArgs("owner-1").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(int64(5)))
	mock.ExpectExec(`INSERT INTO keeper_records`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewRecordRepository(db)
	conflicts, err := repo.UpdateBatch(context.Background(), "owner-1", []*model.Record{
		{ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("client"), Version: 3},
		{ID: "r2", Type: model.RecordTypeText, Ciphertext: []byte("new"), Version: 6},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 1 || conflicts[0].ID != "r1" || conflicts[0].Version != 5 {
		t.Fatalf("unexpected conflicts: %+v", conflicts)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRepositoryUpdateBatchEmptyInput(t *testing.T) {
	t.Parallel()

	repo := NewRecordRepository(nil)
	conflicts, err := repo.UpdateBatch(context.Background(), "owner-1", nil)
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
	if conflicts != nil {
		t.Fatalf("expected nil conflicts, got %+v", conflicts)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	t.Parallel()

	if isUniqueViolation(&pgconn.PgError{Code: "23505"}) != true {
		t.Fatal("expected unique violation")
	}
	if isUniqueViolation(errors.New("other")) != false {
		t.Fatal("expected false for generic error")
	}
}

func TestRecordRepositoryGetNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT id, owner_id, type, meta, ciphertext, version, deleted`).
		WithArgs("missing", "owner-1").
		WillReturnError(sql.ErrNoRows)

	repo := NewRecordRepository(db)
	_, err = repo.Get(context.Background(), "owner-1", "missing")
	if !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

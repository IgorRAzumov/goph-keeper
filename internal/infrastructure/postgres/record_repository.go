package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/record/model"
)

// RecordRepository — PostgreSQL-адаптер для recordrepository.RecordRepository.
type RecordRepository struct {
	db *sql.DB
}

// NewRecordRepository создаёт репозиторий записей.
func NewRecordRepository(db *sql.DB) *RecordRepository {
	return &RecordRepository{db: db}
}

// Get возвращает запись по id для владельца.
func (repository *RecordRepository) Get(ctx context.Context, ownerID, recordID string) (*model.Record, error) {
	if repository == nil || repository.db == nil {
		return nil, common.ErrNotImplemented
	}
	if ownerID == "" || recordID == "" {
		return nil, common.ErrInvalidInput
	}

	row := repository.db.QueryRowContext(ctx, `
SELECT id, owner_id, type, meta, ciphertext, version, deleted
FROM keeper_records
WHERE id=$1 AND owner_id=$2
LIMIT 1
`, recordID, ownerID)

	return scanRecord(row)
}

// ListSince возвращает записи владельца с version > sinceVersion.
func (repository *RecordRepository) ListSince(ctx context.Context, ownerID string, sinceVersion int64) ([]*model.Record, error) {
	if repository == nil || repository.db == nil {
		return nil, common.ErrNotImplemented
	}
	if ownerID == "" || sinceVersion < 0 {
		return nil, common.ErrInvalidInput
	}

	rows, err := repository.db.QueryContext(ctx, `
SELECT id, owner_id, type, meta, ciphertext, version, deleted
FROM keeper_records
WHERE owner_id=$1 AND version > $2
ORDER BY version ASC
`, ownerID, sinceVersion)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []*model.Record
	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if records == nil {
		records = []*model.Record{}
	}
	return records, nil
}

// Update создаёт или обновляет запись (принимает только более новую version).
func (repository *RecordRepository) Update(ctx context.Context, ownerID string, record *model.Record) error {
	if repository == nil || repository.db == nil {
		return common.ErrNotImplemented
	}
	if ownerID == "" || record == nil || record.ID == "" {
		return common.ErrInvalidInput
	}

	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()

	conflict, err := repository.applyRecordInTx(ctx, transaction, ownerID, record)
	if err != nil {
		return err
	}
	if conflict != nil {
		return common.ErrConflict
	}
	return transaction.Commit()
}

// UpdateBatch применяет записи атомарно: при фатальной ошибке откатывает всё; конфликты версий собирает без отката.
func (repository *RecordRepository) UpdateBatch(ctx context.Context, ownerID string, records []*model.Record) ([]*model.Record, error) {
	if repository == nil || repository.db == nil {
		return nil, common.ErrNotImplemented
	}
	if ownerID == "" {
		return nil, common.ErrInvalidInput
	}
	if len(records) == 0 {
		return []*model.Record{}, nil
	}

	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = transaction.Rollback() }()

	var conflicts []*model.Record
	for _, record := range records {
		if record == nil || record.ID == "" {
			return nil, common.ErrInvalidInput
		}
		conflict, err := repository.applyRecordInTx(ctx, transaction, ownerID, record)
		if err != nil {
			return nil, err
		}
		if conflict != nil {
			conflicts = append(conflicts, conflict)
		}
	}
	if err := transaction.Commit(); err != nil {
		return nil, err
	}
	if conflicts == nil {
		conflicts = []*model.Record{}
	}
	return conflicts, nil
}

func (repository *RecordRepository) applyRecordInTx(
	ctx context.Context,
	transaction *sql.Tx,
	ownerID string,
	record *model.Record,
) (*model.Record, error) {
	var storedVersion int64
	err := transaction.QueryRowContext(ctx, `
SELECT version FROM keeper_records WHERE id=$1 AND owner_id=$2
`, record.ID, ownerID).Scan(&storedVersion)
	if errors.Is(err, sql.ErrNoRows) {
		maxVersion, maxErr := maxOwnerVersion(ctx, transaction, ownerID)
		if maxErr != nil {
			return nil, maxErr
		}
		if record.Version <= maxVersion {
			return nil, common.ErrConflict
		}
		_, err = transaction.ExecContext(ctx, `
INSERT INTO keeper_records (id, owner_id, type, meta, ciphertext, version, deleted)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, record.ID, ownerID, string(record.Type), record.Meta, record.Ciphertext, record.Version, record.Deleted)
		if err != nil {
			if isUniqueViolation(err) {
				return nil, common.ErrConflict
			}
			return nil, err
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if record.Version <= storedVersion {
		return getRecordInTx(ctx, transaction, ownerID, record.ID)
	}

	result, err := transaction.ExecContext(ctx, `
UPDATE keeper_records
SET type=$3, meta=$4, ciphertext=$5, version=$6, deleted=$7, updated_at=now()
WHERE id=$1 AND owner_id=$2 AND version < $6
`, record.ID, ownerID, string(record.Type), record.Meta, record.Ciphertext, record.Version, record.Deleted)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return getRecordInTx(ctx, transaction, ownerID, record.ID)
	}
	return nil, nil
}

func maxOwnerVersion(ctx context.Context, transaction *sql.Tx, ownerID string) (int64, error) {
	var maxVersion int64
	err := transaction.QueryRowContext(ctx, `
SELECT COALESCE(MAX(version), 0) FROM keeper_records WHERE owner_id=$1
`, ownerID).Scan(&maxVersion)
	return maxVersion, err
}

func getRecordInTx(ctx context.Context, transaction *sql.Tx, ownerID, recordID string) (*model.Record, error) {
	row := transaction.QueryRowContext(ctx, `
SELECT id, owner_id, type, meta, ciphertext, version, deleted
FROM keeper_records
WHERE id=$1 AND owner_id=$2
LIMIT 1
`, recordID, ownerID)
	return scanRecord(row)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type recordScanner interface {
	Scan(dest ...any) error
}

func scanRecord(row recordScanner) (*model.Record, error) {
	var record model.Record
	var recordType string
	if err := row.Scan(
		&record.ID,
		&record.OwnerID,
		&recordType,
		&record.Meta,
		&record.Ciphertext,
		&record.Version,
		&record.Deleted,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	record.Type = model.RecordType(recordType)
	return &record, nil
}

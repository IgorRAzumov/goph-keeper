package postgres

import (
	"context"
	"database/sql"
	"errors"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/user/model"

	"github.com/jackc/pgx/v5/pgconn"
)

// UserRepository — PostgreSQL-адаптер для domain.UserRepository.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создаёт репозиторий на базе *sql.DB.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Save сохраняет нового пользователя.
func (repository *UserRepository) Save(ctx context.Context, user *model.User) error {
	if repository == nil || repository.db == nil {
		return common.ErrNotImplemented
	}
	if user == nil || user.ID == "" || user.Login == "" || len(user.PasswordHash) == 0 {
		return common.ErrInvalidInput
	}

	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()

	var exists int
	err = transaction.QueryRowContext(ctx, `SELECT 1 FROM keeper_users WHERE login=$1 LIMIT 1`, user.Login).Scan(&exists)
	if err == nil {
		return common.ErrConflict
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	_, err = transaction.ExecContext(ctx, `
INSERT INTO keeper_users (id, login, password_hash)
VALUES ($1, $2, $3)
`, user.ID, user.Login, user.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return common.ErrConflict
		}
		return err
	}
	return transaction.Commit()
}

// GetByLogin получение по логину
func (repository *UserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	if repository == nil || repository.db == nil {
		return nil, common.ErrNotImplemented
	}
	if login == "" {
		return nil, common.ErrInvalidInput
	}

	row := repository.db.QueryRowContext(ctx, `
SELECT id, login, password_hash
FROM keeper_users
WHERE login=$1
LIMIT 1
`, login)

	var user model.User
	if err := row.Scan(&user.ID, &user.Login, &user.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/user/model"

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

	_, err := repository.db.ExecContext(ctx, `
INSERT INTO keeper_users (id, login, password_hash, master_salt)
VALUES ($1, $2, $3, $4)
`, user.ID, user.Login, user.PasswordHash, user.MasterSalt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return common.ErrConflict
		}
		return err
	}
	return nil
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
SELECT id, login, password_hash, master_salt
FROM keeper_users
WHERE login=$1
LIMIT 1
`, login)

	var user model.User
	if err := row.Scan(&user.ID, &user.Login, &user.PasswordHash, &user.MasterSalt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// SetMasterSalt сохраняет master_salt пользователя по id.
func (repository *UserRepository) SetMasterSalt(ctx context.Context, userID, masterSalt string) error {
	if repository == nil || repository.db == nil {
		return common.ErrNotImplemented
	}
	if userID == "" || masterSalt == "" {
		return common.ErrInvalidInput
	}
	result, err := repository.db.ExecContext(ctx, `
UPDATE keeper_users
SET master_salt=$2
WHERE id=$1
`, userID, masterSalt)
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
	return nil
}

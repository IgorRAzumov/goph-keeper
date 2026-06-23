package logout

import (
	"context"
	"errors"
	"strings"
	"time"

	apptoken "goph-keeper/internal/server/application/auth/token"
	"goph-keeper/internal/server/domain/common"
	sessionrepo "goph-keeper/internal/server/domain/session/repository"
)

// Usecase завершает серверную сессию по access JWT (удаляет refresh-сессию по sid из claims).
type Usecase struct {
	sessionRepository sessionrepo.SessionStore
	jwt               apptoken.Provider
}

// NewLogoutUsecase создаёт сценарий выхода.
func NewLogoutUsecase(sessionRepository sessionrepo.SessionStore, jwtProvider apptoken.Provider) *Usecase {
	return &Usecase{
		sessionRepository: sessionRepository,
		jwt:               jwtProvider,
	}
}

func (usecase *Usecase) Execute(ctx context.Context, in Input) error {
	token := strings.TrimSpace(in.AccessToken)
	if token == "" {
		return common.ErrInvalidInput
	}
	if usecase == nil || usecase.sessionRepository == nil || usecase.jwt == nil {
		return common.ErrNotImplemented
	}

	now := time.Now().UTC()
	claims, err := usecase.jwt.ParseAccessToken(token, now)
	if err != nil {
		return common.ErrUnauthorized
	}

	sess, err := usecase.sessionRepository.Get(ctx, claims.Sid)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return err
	}
	if sess.UserID != claims.Sub {
		return common.ErrUnauthorized
	}

	if err := usecase.sessionRepository.Delete(ctx, claims.Sid); err != nil {
		return err
	}
	return nil
}

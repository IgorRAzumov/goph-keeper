package logout

import (
	"context"
	"errors"
	"testing"
	"time"

	"goph-keeper/internal/application/auth/authtest"
	"goph-keeper/internal/domain/common"
	sessionmodel "goph-keeper/internal/domain/session/model"
	sessionmocks "goph-keeper/internal/domain/session/repository/mocks"
	"goph-keeper/internal/security/jwt"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLogoutUsecaseRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionRepository.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)

	usecase := NewLogoutUsecase(sessionRepository, jwt.NewProvider(authtest.UnitTestConfig().JWTSecret))

	err := usecase.Execute(context.Background(), Input{AccessToken: "  "})
	require.ErrorIs(t, err, common.ErrInvalidInput)
}

func TestLogoutUsecaseUnauthorizedOnBadJWT(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionRepository.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)

	usecase := NewLogoutUsecase(sessionRepository, jwt.NewProvider(authtest.UnitTestConfig().JWTSecret))

	err := usecase.Execute(context.Background(), Input{AccessToken: "not-a-jwt"})
	require.ErrorIs(t, err, common.ErrUnauthorized)
}

func TestLogoutUsecaseIdempotentWhenSessionMissing(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	now := time.Now().UTC()
	provider := jwt.NewProvider(authtest.UnitTestConfig().JWTSecret)
	token, err := provider.IssueAccessToken("user-1", "session-1", time.Minute, now)
	require.NoError(t, err)

	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionRepository.EXPECT().
		Get(gomock.Any(), "session-1").
		Return((*sessionmodel.Session)(nil), common.ErrNotFound)

	usecase := NewLogoutUsecase(sessionRepository, provider)
	require.NoError(t, usecase.Execute(context.Background(), Input{AccessToken: token}))
}

func TestLogoutUsecaseUnauthorizedWhenUserMismatch(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	now := time.Now().UTC()
	provider := jwt.NewProvider(authtest.UnitTestConfig().JWTSecret)
	token, err := provider.IssueAccessToken("user-1", "session-1", time.Minute, now)
	require.NoError(t, err)

	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionRepository.EXPECT().
		Get(gomock.Any(), "session-1").
		Return(&sessionmodel.Session{
			ID:               "session-1",
			UserID:           "user-2",
			RefreshTokenHash: sessionmodel.HashRefreshToken("refresh"),
			RefreshExpiresAt: now.Add(time.Hour),
		}, nil)
	sessionRepository.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)

	usecase := NewLogoutUsecase(sessionRepository, provider)
	err = usecase.Execute(context.Background(), Input{AccessToken: token})
	require.ErrorIs(t, err, common.ErrUnauthorized)
}

func TestLogoutUsecaseSuccessDeletesSession(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	now := time.Now().UTC()
	provider := jwt.NewProvider(authtest.UnitTestConfig().JWTSecret)
	token, err := provider.IssueAccessToken("user-1", "session-1", time.Minute, now)
	require.NoError(t, err)

	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionRepository.EXPECT().
		Get(gomock.Any(), "session-1").
		Return(&sessionmodel.Session{
			ID:               "session-1",
			UserID:           "user-1",
			RefreshTokenHash: sessionmodel.HashRefreshToken("refresh"),
			RefreshExpiresAt: now.Add(time.Hour),
		}, nil)
	sessionRepository.EXPECT().
		Delete(gomock.Any(), "session-1").
		Return(nil)

	usecase := NewLogoutUsecase(sessionRepository, provider)
	require.NoError(t, usecase.Execute(context.Background(), Input{AccessToken: token}))
}

func TestLogoutUsecasePropagatesRepositoryErrors(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	now := time.Now().UTC()
	provider := jwt.NewProvider(authtest.UnitTestConfig().JWTSecret)
	token, err := provider.IssueAccessToken("user-1", "session-1", time.Minute, now)
	require.NoError(t, err)

	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionRepository.EXPECT().
		Get(gomock.Any(), "session-1").
		Return((*sessionmodel.Session)(nil), errors.New("db down"))

	usecase := NewLogoutUsecase(sessionRepository, provider)
	err = usecase.Execute(context.Background(), Input{AccessToken: token})
	require.Error(t, err)
}

package refresh

import (
	"context"
	"testing"
	"time"

	"goph-keeper/internal/server/application/auth/authtest"
	"goph-keeper/internal/server/domain/common"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
	sessionmocks "goph-keeper/internal/server/domain/session/repository/mocks"
	sessionsvc "goph-keeper/internal/server/domain/session/service"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRefreshUsecaseRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	sessionRepository := sessionmocks.NewMockSessionStore(controller)
	sessionService := sessionsvc.NewSessionService(sessionRepository)
	usecase := NewRefreshUsecase(sessionRepository, sessionService, authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL)

	sessionRepository.EXPECT().FindActiveByRefreshToken(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	_, err := usecase.Execute(context.Background(), Input{RefreshToken: "  "})
	require.ErrorIs(t, err, common.ErrInvalidInput)
}

func TestRefreshUsecaseUnauthorizedWhenSessionNotFound(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	sessionRepository := sessionmocks.NewMockSessionStore(controller)
	sessionRepository.EXPECT().
		FindActiveByRefreshToken(gomock.Any(), "refresh-1", gomock.AssignableToTypeOf(time.Time{})).
		Return((*sessionmodel.Session)(nil), common.ErrNotFound)

	sessionService := sessionsvc.NewSessionService(sessionRepository)
	usecase := NewRefreshUsecase(sessionRepository, sessionService, authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL)

	_, err := usecase.Execute(context.Background(), Input{RefreshToken: "refresh-1"})
	require.ErrorIs(t, err, common.ErrUnauthorized)
}

func TestRefreshUsecaseUnauthorizedWhenRotateNotFound(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	now := time.Now().UTC()
	session := &sessionmodel.Session{
		ID:               "session-old",
		UserID:           "user-1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("refresh-1"),
		RefreshExpiresAt: now.Add(time.Hour),
	}

	sessionRepository := sessionmocks.NewMockSessionStore(controller)
	sessionRepository.EXPECT().
		FindActiveByRefreshToken(gomock.Any(), "refresh-1", gomock.AssignableToTypeOf(time.Time{})).
		Return(session, nil)
	sessionRepository.EXPECT().
		Rotate(gomock.Any(), "session-old", gomock.Cond(func(candidate any) bool {
			newSession, ok := candidate.(*sessionmodel.Session)
			return ok && newSession != nil
		})).
		Return(common.ErrNotFound)

	sessionService := sessionsvc.NewSessionService(sessionRepository)
	usecase := NewRefreshUsecase(sessionRepository, sessionService, authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL)

	_, err := usecase.Execute(context.Background(), Input{RefreshToken: "refresh-1"})
	require.ErrorIs(t, err, common.ErrUnauthorized)
}

func TestRefreshUsecaseSuccessRotatesSession(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	now := time.Now().UTC()
	session := &sessionmodel.Session{
		ID:               "session-old",
		UserID:           "user-1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("refresh-1"),
		RefreshExpiresAt: now.Add(time.Hour),
	}

	sessionRepository := sessionmocks.NewMockSessionStore(controller)
	sessionRepository.EXPECT().
		FindActiveByRefreshToken(gomock.Any(), "refresh-1", gomock.AssignableToTypeOf(time.Time{})).
		Return(session, nil)
	sessionRepository.EXPECT().
		Rotate(gomock.Any(), "session-old", gomock.Cond(func(candidate any) bool {
			newSession, ok := candidate.(*sessionmodel.Session)
			if !ok || newSession == nil {
				return false
			}
			if newSession.ID == "" || newSession.ID == "session-old" {
				return false
			}
			if newSession.UserID != "user-1" {
				return false
			}
			if len(newSession.RefreshTokenHash) != len(sessionmodel.HashRefreshToken("")) {
				return false
			}
			return newSession.RefreshExpiresAt.After(now)
		})).
		Return(nil)

	sessionService := sessionsvc.NewSessionService(sessionRepository)
	usecase := NewRefreshUsecase(sessionRepository, sessionService, authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL)

	output, err := usecase.Execute(context.Background(), Input{RefreshToken: "refresh-1"})
	require.NoError(t, err)
	require.NotEmpty(t, output.AccessToken)
	require.NotEmpty(t, output.RefreshToken)

	provider := authtest.Provider()
	claims, err := provider.ParseAccessToken(output.AccessToken, now)
	require.NoError(t, err)
	require.Equal(t, "user-1", claims.Sub)
	require.NotEmpty(t, claims.Sid)
}

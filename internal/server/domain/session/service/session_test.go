package service

import (
	"context"
	"testing"
	"time"

	sessionmodel "goph-keeper/internal/server/domain/session/model"
	sessionmocks "goph-keeper/internal/server/domain/session/repository/mocks"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateSessionStoresDeterministicRefreshHash(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	repo := sessionmocks.NewMockSessionStore(ctrl)

	var saved *sessionmodel.Session
	repo.EXPECT().
		Save(gomock.Any(), gomock.AssignableToTypeOf(&sessionmodel.Session{})).
		DoAndReturn(func(_ context.Context, s *sessionmodel.Session) error {
			c := *s
			c.RefreshTokenHash = append([]byte(nil), s.RefreshTokenHash...)
			saved = &c
			return nil
		})

	service := NewSessionService(repo)
	now := time.Now().UTC()

	sessionID, err := service.CreateSession(context.Background(), "user-1", "refresh-token", time.Hour, now)
	require.NoError(t, err)

	repo.EXPECT().Get(gomock.Any(), sessionID).Return(saved, nil)
	session, err := repo.Get(context.Background(), sessionID)
	require.NoError(t, err)
	require.Equal(t, string(sessionmodel.HashRefreshToken("refresh-token")), string(session.RefreshTokenHash))
}

func TestRotateSessionUsesRepositoryAtomicRotate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	repo := sessionmocks.NewMockSessionStore(ctrl)

	repo.EXPECT().
		Save(gomock.Any(), gomock.AssignableToTypeOf(&sessionmodel.Session{})).
		Return(nil)

	service := NewSessionService(repo)
	now := time.Now().UTC()

	oldSessionID, err := service.CreateSession(context.Background(), "user-1", "old-refresh", time.Hour, now)
	require.NoError(t, err)

	repo.EXPECT().
		Rotate(gomock.Any(), oldSessionID, gomock.AssignableToTypeOf(&sessionmodel.Session{})).
		DoAndReturn(func(_ context.Context, oldID string, newSess *sessionmodel.Session) error {
			require.Equal(t, oldSessionID, oldID)
			require.NotEmpty(t, newSess.ID)
			require.NotEqual(t, oldSessionID, newSess.ID)
			require.Equal(t, "user-1", newSess.UserID)
			return nil
		})

	newSessionID, err := service.RotateSession(context.Background(), oldSessionID, "user-1", "new-refresh", time.Hour, now)
	require.NoError(t, err)
	require.NotEqual(t, oldSessionID, newSessionID)
}

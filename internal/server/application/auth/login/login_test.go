package login

import (
	"context"
	"testing"
	"time"

	"goph-keeper/internal/server/application/auth/authtest"
	"goph-keeper/internal/server/domain/common"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
	sessionmocks "goph-keeper/internal/server/domain/session/repository/mocks"
	sessionsvc "goph-keeper/internal/server/domain/session/service"
	"goph-keeper/internal/server/domain/user/model"
	usermocks "goph-keeper/internal/server/domain/user/repository/mocks"
	usersvc "goph-keeper/internal/server/domain/user/service"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginUsecaseRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	userRepository := usermocks.NewMockUserRepository(controller)
	userService := usersvc.NewUserService(userRepository)
	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionService := sessionsvc.NewSessionService(sessionRepository)

	sessionRepository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)

	usecase := NewLoginUsecase(userService, sessionService, authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL)

	_, err := usecase.Execute(context.Background(), Input{Login: " ", Password: ""})
	require.ErrorIs(t, err, common.ErrInvalidInput)
}

func TestLoginUsecaseSuccessPersistsSession(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	t.Cleanup(controller.Finish)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	require.NoError(t, err)

	userRepository := usermocks.NewMockUserRepository(controller)
	userRepository.EXPECT().
		GetByLogin(gomock.Any(), "alice").
		Return(&model.User{
			ID:           "user-1",
			Login:        "alice",
			PasswordHash: passwordHash,
			MasterSalt:   "salt-1",
		}, nil)

	userService := usersvc.NewUserService(userRepository)

	var savedSession *sessionmodel.Session
	sessionRepository := sessionmocks.NewMockSessionRepository(controller)
	sessionRepository.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, session *sessionmodel.Session) error {
			savedSession = session
			return nil
		})

	sessionService := sessionsvc.NewSessionService(sessionRepository)
	usecase := NewLoginUsecase(userService, sessionService, authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL)

	output, err := usecase.Execute(context.Background(), Input{Login: "alice", Password: "secret"})
	require.NoError(t, err)
	require.NotEmpty(t, output.AccessToken)
	require.NotEmpty(t, output.RefreshToken)

	require.NotNil(t, savedSession)
	require.Equal(t, "user-1", savedSession.UserID)
	require.NotEmpty(t, savedSession.ID)
	require.Equal(t, string(sessionmodel.HashRefreshToken(output.RefreshToken)), string(savedSession.RefreshTokenHash))
	require.True(t, savedSession.RefreshExpiresAt.After(time.Now().UTC()))
}

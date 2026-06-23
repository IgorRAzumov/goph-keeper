package testserver

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"goph-keeper/internal/server/application/auth/authtest"
	applogin "goph-keeper/internal/server/application/auth/login"
	applogout "goph-keeper/internal/server/application/auth/logout"
	apprefresh "goph-keeper/internal/server/application/auth/refresh"
	appregister "goph-keeper/internal/server/application/auth/register"
	appverify "goph-keeper/internal/server/application/auth/verify"
	appsync "goph-keeper/internal/server/application/sync"
	httpapi "goph-keeper/internal/server/delivery/http"
	recordsvc "goph-keeper/internal/server/domain/record/service"
	sessionrepo "goph-keeper/internal/server/domain/session/repository"
	sessionsvc "goph-keeper/internal/server/domain/session/service"
	userrepo "goph-keeper/internal/server/domain/user/repository"
	usersvc "goph-keeper/internal/server/domain/user/service"
	"goph-keeper/internal/server/security/jwt"
	"goph-keeper/internal/testutil/fake"
)

// Fixture — тестовый HTTP-сервер с in-memory зависимостями.
type Fixture struct {
	Server   *httptest.Server
	Users    userrepo.UserStore
	Sessions sessionrepo.SessionStore
	Records  *fake.RecordRepo
	JWT      *jwt.Provider
}

// New создаёт httptest.Server с полным API-стеком на fake-репозиториях.
func New(t *testing.T) *Fixture {
	t.Helper()

	users := fake.NewUserRepo()
	sessions := fake.NewSessionRepo()
	records := fake.NewRecordRepo()

	userService := usersvc.NewUserService(users)
	sessionService := sessionsvc.NewSessionService(sessions)
	recordService := recordsvc.NewRecordService(records)
	jwtProvider := authtest.Provider()

	deps := httpapi.Dependencies{
		RegisterUser: appregister.NewUserUsecase(userService),
		Login:        applogin.NewLoginUsecase(userService, sessionService, jwtProvider, authtest.AccessTTL, authtest.RefreshTTL),
		Refresh:      apprefresh.NewRefreshUsecase(sessions, sessionService, jwtProvider, authtest.AccessTTL, authtest.RefreshTTL),
		Logout:       applogout.NewLogoutUsecase(sessions, jwtProvider),
		Verify:       appverify.NewUsecase(jwtProvider, sessions),
		Sync:         appsync.NewUsecase(recordService),
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(httpapi.Router(logger, deps))

	t.Cleanup(server.Close)

	return &Fixture{
		Server:   server,
		Users:    users,
		Sessions: sessions,
		Records:  records,
		JWT:      jwtProvider,
	}
}

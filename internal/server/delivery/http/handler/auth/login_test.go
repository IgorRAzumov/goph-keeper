package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goph-keeper/internal/server/application/auth/authtest"
	applogin "goph-keeper/internal/server/application/auth/login"
	applogout "goph-keeper/internal/server/application/auth/logout"
	apprefresh "goph-keeper/internal/server/application/auth/refresh"
	"goph-keeper/internal/server/domain/common"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
	sessionmocks "goph-keeper/internal/server/domain/session/repository/mocks"
	sessionsvc "goph-keeper/internal/server/domain/session/service"
	"goph-keeper/internal/server/domain/user/model"
	usermocks "goph-keeper/internal/server/domain/user/repository/mocks"
	usersvc "goph-keeper/internal/server/domain/user/service"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginReturnsTokens(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	userRepo := usermocks.NewMockUserStore(ctrl)
	userRepo.EXPECT().GetByLogin(gomock.Any(), "alice").Return(&model.User{
		ID: "u1", Login: "alice", PasswordHash: hash,
		MasterSalt: "salt-1",
	}, nil)

	sessionRepo := sessionmocks.NewMockSessionStore(ctrl)
	sessionRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	handler := Login(testLogger(), applogin.NewLoginUsecase(usersvc.NewUserService(userRepo), sessionsvc.NewSessionService(sessionRepo), authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL))

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"login":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["access_token"] == "" || body["refresh_token"] == "" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	jwtProvider := authtest.Provider()
	token, err := jwtProvider.IssueAccessToken("u1", "s1", time.Minute, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	sessionRepo := sessionmocks.NewMockSessionStore(ctrl)
	sessionRepo.EXPECT().Get(gomock.Any(), "s1").Return(&sessionmodel.Session{
		ID: "s1", UserID: "u1",
	}, nil)
	sessionRepo.EXPECT().Delete(gomock.Any(), "s1").Return(nil)

	handler := Logout(testLogger(), applogout.NewLogoutUsecase(sessionRepo, jwtProvider))
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRefreshReturnsNewTokens(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	now := time.Now().UTC()
	sessionRepo := sessionmocks.NewMockSessionStore(ctrl)
	sessionRepo.EXPECT().FindActiveByRefreshToken(gomock.Any(), "old-refresh", gomock.Any()).
		Return(&sessionmodel.Session{
			ID: "s1", UserID: "u1",
			RefreshTokenHash: sessionmodel.HashRefreshToken("old-refresh"),
			RefreshExpiresAt: now.Add(time.Hour),
		}, nil)
	sessionRepo.EXPECT().Rotate(gomock.Any(), "s1", gomock.Any()).Return(nil)

	handler := Refresh(testLogger(), apprefresh.NewRefreshUsecase(sessionRepo, sessionsvc.NewSessionService(sessionRepo), authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL))

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(`{"refresh_token":"old-refresh"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestLoginRejectsMissingUseCase(t *testing.T) {
	t.Parallel()

	handler := Login(testLogger(), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
}

func TestLogoutRejectsMissingToken(t *testing.T) {
	t.Parallel()

	handler := Logout(testLogger(), applogout.NewLogoutUsecase(sessionmocks.NewMockSessionStore(gomock.NewController(t)), authtest.Provider()))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/logout", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRefreshRejectsEmptyBody(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	sessionRepo := sessionmocks.NewMockSessionStore(ctrl)
	handler := Refresh(testLogger(), apprefresh.NewRefreshUsecase(sessionRepo, sessionsvc.NewSessionService(sessionRepo), authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/refresh", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRefreshMapsUnauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	sessionRepo := sessionmocks.NewMockSessionStore(ctrl)
	sessionRepo.EXPECT().FindActiveByRefreshToken(gomock.Any(), "bad", gomock.Any()).Return(nil, common.ErrNotFound)

	handler := Refresh(testLogger(), apprefresh.NewRefreshUsecase(sessionRepo, sessionsvc.NewSessionService(sessionRepo), authtest.Provider(), authtest.AccessTTL, authtest.RefreshTTL))
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(`{"refresh_token":"bad"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

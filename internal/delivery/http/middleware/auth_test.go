package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goph-keeper/internal/domain/common"
	sessionmodel "goph-keeper/internal/domain/session/model"
	sessionmocks "goph-keeper/internal/domain/session/repository/mocks"
	"goph-keeper/internal/security/jwt"

	"go.uber.org/mock/gomock"
)

func TestBearerAuthRejectsRevokedSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	provider := jwt.NewProvider("secret")
	token := mustIssueToken(t, provider, "user-1", "session-1")

	sessionRepository := sessionmocks.NewMockSessionRepository(ctrl)
	sessionRepository.EXPECT().Get(gomock.Any(), "session-1").Return(nil, common.ErrNotFound)

	handler := BearerAuth(provider, sessionRepository)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/sync", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestBearerAuthAcceptsExistingSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	provider := jwt.NewProvider("secret")
	sessionRepository := sessionmocks.NewMockSessionRepository(ctrl)
	sessionRepository.EXPECT().Get(gomock.Any(), "session-1").Return(&sessionmodel.Session{
		ID:               "session-1",
		UserID:           "user-1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("refresh-token"),
		RefreshExpiresAt: time.Now().UTC().Add(time.Hour),
	}, nil)

	token := mustIssueToken(t, provider, "user-1", "session-1")
	handler := BearerAuth(provider, sessionRepository)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if UserID(request.Context()) != "user-1" || SessionID(request.Context()) != "session-1" {
			t.Fatalf("auth context was not populated")
		}
		writer.WriteHeader(http.StatusAccepted)
	}))

	request := httptest.NewRequest(http.MethodGet, "/sync", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, response.Code)
	}
}

func mustIssueToken(t *testing.T, provider *jwt.Provider, userID, sessionID string) string {
	t.Helper()
	token, err := provider.IssueAccessToken(userID, sessionID, time.Minute, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

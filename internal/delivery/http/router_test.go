package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sessionmodel "goph-keeper/internal/domain/session/model"
	sessionmocks "goph-keeper/internal/domain/session/repository/mocks"
	"goph-keeper/internal/security/jwt"

	"go.uber.org/mock/gomock"
)

func TestRouterHealthz(t *testing.T) {
	router := Router(nil, Dependencies{})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body struct {
		Status    string `json:"status"`
		Version   string `json:"version"`
		BuildDate string `json:"buildDate"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("unexpected health status: %q", body.Status)
	}
}

func TestRouterAuthRegisterWithoutUseCaseReturnsNotImplemented(t *testing.T) {
	router := Router(nil, Dependencies{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, response.Code)
	}
}

func TestRouterSyncRequiresBearerThenStub(t *testing.T) {
	provider := jwt.NewProvider("router-test-secret")

	tests := []struct {
		name         string
		method       string
		path         string
		authHeader   string
		wantStatus   int
		wantStubBody bool
	}{
		{
			name:       "no auth",
			method:     http.MethodGet,
			path:       "/api/v1/sync/",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			method:     http.MethodGet,
			path:       "/api/v1/sync/",
			authHeader: "Bearer not-a-jwt",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "valid token hits stub",
			method:       http.MethodGet,
			path:         "/api/v1/sync/",
			authHeader:   "Bearer " + mustAccessToken(t, provider),
			wantStatus:   http.StatusNotImplemented,
			wantStubBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			sessions := sessionmocks.NewMockSessionRepository(ctrl)
			if tt.wantStubBody {
				sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(&sessionmodel.Session{
					ID:               "sess-1",
					UserID:           "user-1",
					RefreshTokenHash: sessionmodel.HashRefreshToken("refresh"),
					RefreshExpiresAt: time.Now().UTC().Add(time.Hour),
				}, nil)
			}

			deps := Dependencies{JWT: provider, Sessions: sessions}
			router := Router(nil, deps)

			request := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				request.Header.Set("Authorization", tt.authHeader)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, response.Code)
			}
			if tt.wantStubBody {
				var body map[string]any
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if body["error"] != "not implemented" {
					t.Fatalf("unexpected body: %#v", body)
				}
			}
		})
	}
}

func mustAccessToken(t *testing.T, p *jwt.Provider) string {
	t.Helper()
	token, err := p.IssueAccessToken("user-1", "sess-1", time.Minute, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

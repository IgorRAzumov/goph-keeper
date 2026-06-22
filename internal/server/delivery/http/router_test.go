package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appverify "goph-keeper/internal/server/application/auth/verify"
	appsync "goph-keeper/internal/server/application/sync"
	"goph-keeper/internal/server/domain/record/model"
	repomocks "goph-keeper/internal/server/domain/record/repository/mocks"
	recordsvc "goph-keeper/internal/server/domain/record/service"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
	sessionmocks "goph-keeper/internal/server/domain/session/repository/mocks"
	"goph-keeper/internal/server/security/jwt"

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

func TestRouterSyncRequiresBearer(t *testing.T) {
	provider := jwt.NewProvider("router-test-secret")

	tests := []struct {
		name       string
		method     string
		path       string
		authHeader string
		wantStatus int
		setupSync  bool
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
			name:       "valid token pull",
			method:     http.MethodGet,
			path:       "/api/v1/sync/?since=0",
			authHeader: "Bearer " + mustAccessToken(t, provider),
			wantStatus: http.StatusOK,
			setupSync:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			sessions := sessionmocks.NewMockSessionRepository(ctrl)
			if tt.setupSync {
				sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(&sessionmodel.Session{
					ID:               "sess-1",
					UserID:           "user-1",
					RefreshTokenHash: sessionmodel.HashRefreshToken("refresh"),
					RefreshExpiresAt: time.Now().UTC().Add(time.Hour),
				}, nil)
			}

			var syncUsecase *appsync.Usecase
			if tt.setupSync {
				records := repomocks.NewMockRecordRepository(ctrl)
				records.EXPECT().ListSince(gomock.Any(), "user-1", int64(0)).Return([]*model.Record{}, nil)
				syncUsecase = appsync.NewUsecase(recordsvc.NewRecordService(records))
			}

			deps := Dependencies{Verify: appverify.NewUsecase(provider, sessions), Sync: syncUsecase}
			router := Router(nil, deps)

			request := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				request.Header.Set("Authorization", tt.authHeader)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d body=%s", tt.wantStatus, response.Code, response.Body.String())
			}
			if tt.setupSync {
				var body map[string]any
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if _, ok := body["records"]; !ok {
					t.Fatalf("expected records in body: %#v", body)
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

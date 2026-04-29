package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	appuser "goph-keeper/internal/application/user"
	usersvc "goph-keeper/internal/domain/user/service"
	"goph-keeper/internal/infrastructure/memory"
	"goph-keeper/internal/logging"
)

func TestAuthRegisterCreatesUser(t *testing.T) {
	t.Parallel()

	handler := Register(testLogger(), newRegisterUserUseCase())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{
		"login": "alice",
		"password": "secret"
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	var body struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.UserID == "" {
		t.Fatal("expected user id")
	}
}

func TestAuthRegisterWithoutUseCaseReturnsNotImplemented(t *testing.T) {
	t.Parallel()

	handler := Register(testLogger(), nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, response.Code)
	}
}

func TestAuthRegisterRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	handler := Register(testLogger(), newRegisterUserUseCase())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestAuthRegisterRequiresPassword(t *testing.T) {
	t.Parallel()

	handler := Register(testLogger(), newRegisterUserUseCase())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{
		"login": "alice"
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestAuthRegisterReturnsConflictForDuplicateLogin(t *testing.T) {
	t.Parallel()

	handler := Register(testLogger(), newRegisterUserUseCase())
	requestBody := []byte(`{"login":"alice","password":"secret"}`)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(requestBody)))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(requestBody)))

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}
}

func newRegisterUserUseCase() *appuser.Usecase {
	repo := memory.NewUserRepository()
	service := usersvc.NewUserService(repo)
	return appuser.NewUserUsecase(service)
}

func testLogger() logging.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

package handler

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
)

func TestAuthRegisterCreatesUser(t *testing.T) {
	t.Parallel()

	handler := AuthRegister(testLogger(), newRegisterUserUseCase())
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

func TestAuthRegisterRequiresPassword(t *testing.T) {
	t.Parallel()

	handler := AuthRegister(testLogger(), newRegisterUserUseCase())
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

	handler := AuthRegister(testLogger(), newRegisterUserUseCase())
	requestBody := []byte(`{"login":"alice","password":"secret"}`)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(requestBody)))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(requestBody)))

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, response.Code)
	}
}

func newRegisterUserUseCase() *appuser.RegisterUserUseCase {
	repo := memory.NewUserRepository()
	service := usersvc.NewUserService(repo)
	return appuser.NewRegisterUserUseCase(service)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

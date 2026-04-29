package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestRouterSyncEndpointsAreStubs(t *testing.T) {
	router := Router(nil, Dependencies{})

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "get sync", method: http.MethodGet, path: "/api/v1/sync/"},
		{name: "post sync", method: http.MethodPost, path: "/api/v1/sync/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusNotImplemented {
				t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, response.Code)
			}
		})
	}
}

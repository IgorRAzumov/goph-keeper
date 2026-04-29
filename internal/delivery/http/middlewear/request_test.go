package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type loggerStub struct {
	infoCalls int
	lastMsg   string
	lastArgs  []any
}

func (l *loggerStub) Info(msg string, args ...any) {
	l.infoCalls++
	l.lastMsg = msg
	l.lastArgs = args
}

func (l *loggerStub) Error(string, ...any) {}

func TestRequestLogMiddlewareLogsRequest(t *testing.T) {
	t.Parallel()

	log := &loggerStub{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	handler := RequestLogMiddleware(log, next)

	request := httptest.NewRequest(http.MethodPatch, "/records/1", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, response.Code)
	}
	if log.infoCalls != 1 {
		t.Fatalf("expected one info log call, got %d", log.infoCalls)
	}
	if log.lastMsg != "http_request" {
		t.Fatalf("unexpected log message: %q", log.lastMsg)
	}
	assertLogArg(t, log.lastArgs, "method", http.MethodPatch)
	assertLogArg(t, log.lastArgs, "path", "/records/1")
}

func TestRequestLogMiddlewareAcceptsNilLogger(t *testing.T) {
	t.Parallel()

	handler := RequestLogMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}

func assertLogArg(t *testing.T, args []any, key string, want any) {
	t.Helper()
	for i := 0; i < len(args)-1; i += 2 {
		if args[i] == key && args[i+1] == want {
			return
		}
	}
	t.Fatalf("expected log arg %q=%v in %#v", key, want, args)
}

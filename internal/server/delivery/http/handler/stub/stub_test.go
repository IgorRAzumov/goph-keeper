package stub

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotImplemented(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	NotImplemented(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
}

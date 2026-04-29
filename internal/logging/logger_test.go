package logging

import "testing"

func TestDefaultReturnsLogger(t *testing.T) {
	t.Parallel()

	log := Default()
	if log == nil {
		t.Fatal("expected default logger")
	}

	log.Info("test_info", "key", "value")
	log.Error("test_error", "key", "value")
}

package buildinfo

import "testing"

func TestBuildInfoDefaults(t *testing.T) {
	t.Parallel()

	if Version == "" {
		t.Fatal("expected version")
	}
	if BuildDate == "" {
		t.Fatal("expected build date")
	}
}

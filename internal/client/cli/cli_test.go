package cli_test

import (
	"testing"

	"goph-keeper/internal/client/cli"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()

	if code := cli.Run([]string{"version"}); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	if code := cli.Run([]string{"nope"}); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()

	if code := cli.Run([]string{"help"}); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
}

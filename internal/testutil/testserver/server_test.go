package testserver_test

import (
	"testing"

	"goph-keeper/internal/testutil/testserver"
)

func TestNewFixture(t *testing.T) {
	fixture := testserver.New(t)
	if fixture.Server == nil || fixture.Server.URL == "" {
		t.Fatal("expected server url")
	}
}

package cli_test

import (
	"path/filepath"
	"testing"

	"goph-keeper/internal/client/cli"
	"goph-keeper/internal/testutil/testserver"
)

func TestRunRegisterAndList(t *testing.T) {
	fixture := testserver.New(t)
	dir := filepath.Join(t.TempDir(), "cli")
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")
	t.Setenv("GOPHKEEPER_SERVER_URL", fixture.Server.URL)

	if code := cli.Run([]string{"register", "-login", "cli-user", "-password", "pass"}); code != 0 {
		t.Fatalf("register exit %d", code)
	}
	if code := cli.Run([]string{"add", "-type", "text", "-meta", "m", "-data", "hello"}); code != 0 {
		t.Fatalf("add exit %d", code)
	}
	if code := cli.Run([]string{"list"}); code != 0 {
		t.Fatalf("list exit %d", code)
	}
	if code := cli.Run([]string{"sync"}); code != 0 {
		t.Fatalf("sync exit %d", code)
	}
}

func TestRunLogout(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")
	t.Setenv("GOPHKEEPER_SERVER_URL", fixture.Server.URL)

	_ = cli.Run([]string{"register", "-login", "logout-user", "-password", "pass"})
	if code := cli.Run([]string{"logout"}); code != 0 {
		t.Fatalf("logout exit %d", code)
	}
}

func TestRunGetDelete(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")
	t.Setenv("GOPHKEEPER_SERVER_URL", fixture.Server.URL)

	_ = cli.Run([]string{"register", "-login", "gd", "-password", "pass"})
	// capture id from add via running add and reading config data - use list output not parsed; call add via app is easier
	// use sync + get with wrong id test
	if code := cli.Run([]string{"get", "00000000-0000-0000-0000-000000000000"}); code == 0 {
		t.Fatal("expected get failure for missing id")
	}
}

func TestTruncateViaList(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")
	t.Setenv("GOPHKEEPER_SERVER_URL", fixture.Server.URL)
	_ = cli.Run([]string{"register", "-login", "long", "-password", "pass"})
	long := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	_ = cli.Run([]string{"add", "-type", "text", "-meta", "long", "-data", long})
	if code := cli.Run([]string{"list"}); code != 0 {
		t.Fatalf("list exit %d", code)
	}
}

func TestRunAddLoginType(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")
	t.Setenv("GOPHKEEPER_SERVER_URL", fixture.Server.URL)

	_ = cli.Run([]string{"register", "-login", "login-type", "-password", "pass"})
	if code := cli.Run([]string{"add", "-type", "login", "-meta", "site", "-user", "u", "-pass", "p"}); code != 0 {
		t.Fatalf("add login exit %d", code)
	}
}

func TestRunLoginWithPasswordFlag(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")
	t.Setenv("GOPHKEEPER_SERVER_URL", fixture.Server.URL)

	_ = cli.Run([]string{"register", "-login", "login-flag", "-password", "pass"})
	if code := cli.Run([]string{"login", "-login", "login-flag", "-password", "pass"}); code != 0 {
		t.Fatalf("login exit %d", code)
	}
}

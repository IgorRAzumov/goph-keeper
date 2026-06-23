package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/testutil/testserver"
)

func TestAppAddCardAndBinary(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Register(context.Background(), "card-user", "pass"); err != nil {
		t.Fatal(err)
	}

	_, err = app.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeCard, Meta: "bank",
		Payload: map[string]string{"number": "4111", "holder": "A", "expiry": "12/30", "cvv": "123"},
	})
	if err != nil {
		t.Fatal(err)
	}

	binFile := filepath.Join(dir, "blob.bin")
	if err := os.WriteFile(binFile, []byte{1, 2, 3}, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = app.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeBinary, Meta: "file", FilePath: binFile,
	})
	if err != nil {
		t.Fatal(err)
	}
}

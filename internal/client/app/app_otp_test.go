package app_test

import (
	"context"
	"testing"

	clientapp "goph-keeper/internal/client/app"
	"goph-keeper/internal/client/model"
	"goph-keeper/internal/testutil/testserver"
)

func TestAppAddOTPType(t *testing.T) {
	fixture := testserver.New(t)
	dir := t.TempDir()
	t.Setenv("GOPHKEEPER_CONFIG_DIR", dir)
	t.Setenv("GOPHKEEPER_MASTER_PASSWORD", "master")

	app, err := clientapp.New("master", fixture.Server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = app.Register(context.Background(), "otp-user", "pass")
	_, err = app.Add(context.Background(), clientapp.AddInput{
		Type: model.RecordTypeOTP, Meta: "svc", Payload: map[string]string{"secret": "ABC"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

package app

import (
	"context"
	"testing"
	"time"

	"goph-keeper/internal/config"
)

type loggerStub struct {
	infoCalls  int
	errorCalls int
}

func (l *loggerStub) Info(string, ...any) {
	l.infoCalls++
}

func (l *loggerStub) Error(string, ...any) {
	l.errorCalls++
}

func TestNewWithBuildsApp(t *testing.T) {
	t.Parallel()

	log := &loggerStub{}
	app, err := NewWith(config.Config{HTTPAddr: "127.0.0.1:0"}, log)
	if err != nil {
		t.Fatalf("NewWith failed: %v", err)
	}
	defer func() {
		if err := app.httpServer.Run(cancelledContext()); err != nil {
			t.Fatalf("shutdown server: %v", err)
		}
	}()

	if app.httpServer == nil {
		t.Fatal("expected HTTP server")
	}
	if app.Logger() != log {
		t.Fatal("expected injected logger")
	}
	if log.infoCalls != 1 {
		t.Fatalf("expected server listening log call, got %d", log.infoCalls)
	}
}

func TestLoggerReturnsDefaultForNilApp(t *testing.T) {
	t.Parallel()

	var app *App
	if app.Logger() == nil {
		t.Fatal("expected default logger")
	}
}

func TestRunRejectsUninitializedApp(t *testing.T) {
	t.Parallel()

	err := (&App{}).Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	cancel()
	return ctx
}

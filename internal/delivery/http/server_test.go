package httpapi

import (
	"context"
	"testing"
	"time"
)

func TestNewServerUsesConfiguredAddress(t *testing.T) {
	server, err := NewServer(ServerConfig{Address: "127.0.0.1:0"}, nil)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer closeServer(t, server)

	if server.httpServer.Addr != "127.0.0.1:0" {
		t.Fatalf("expected configured addr, got %q", server.httpServer.Addr)
	}
	if server.Addr() == "" {
		t.Fatal("expected listener addr")
	}
}

func TestServerRunShutsDownOnContextCancel(t *testing.T) {
	server, err := NewServer(ServerConfig{Address: "127.0.0.1:0"}, nil)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run(ctx)
	}()

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not stop after context cancel")
	}
}

func TestServerAddrReturnsEmptyForNilListener(t *testing.T) {
	server := &Server{}

	if got := server.Addr(); got != "" {
		t.Fatalf("expected empty addr, got %q", got)
	}
}

func TestNewServerReturnsListenError(t *testing.T) {
	first, err := NewServer(ServerConfig{Address: "127.0.0.1:0"}, nil)
	if err != nil {
		t.Fatalf("first NewServer failed: %v", err)
	}
	defer closeServer(t, first)

	_, err = NewServer(ServerConfig{Address: first.Addr()}, nil)
	if err == nil {
		t.Fatal("expected listen error")
	}
}

func TestNewServerConfiguresHeaderTimeout(t *testing.T) {
	server, err := NewServer(ServerConfig{Address: "127.0.0.1:0"}, nil)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	defer closeServer(t, server)

	if server.httpServer.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("unexpected ReadHeaderTimeout: %s", server.httpServer.ReadHeaderTimeout)
	}
	if server.httpServer.Handler == nil {
		t.Fatal("expected handler")
	}
}

func closeServer(t *testing.T, server *Server) {
	t.Helper()
	if server == nil || server.listener == nil {
		return
	}
	if err := server.listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
}

package verify

import (
	"context"
	"errors"
	"testing"
	"time"

	"goph-keeper/internal/server/application/auth/authtest"
	"goph-keeper/internal/server/domain/common"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
	sessionmocks "goph-keeper/internal/server/domain/session/repository/mocks"

	"go.uber.org/mock/gomock"
)

func issueToken(t *testing.T, userID, sessionID string, now time.Time) string {
	t.Helper()
	token, err := authtest.Provider().IssueAccessToken(userID, sessionID, time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestVerifySuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	now := time.Now().UTC()
	sessions := sessionmocks.NewMockSessionStore(ctrl)
	sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(&sessionmodel.Session{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshExpiresAt: now.Add(time.Hour),
	}, nil)

	usecase := NewUsecase(authtest.Provider(), sessions)
	out, err := usecase.Execute(context.Background(), issueToken(t, "user-1", "sess-1", now), now)
	if err != nil {
		t.Fatal(err)
	}
	if out.UserID != "user-1" || out.SessionID != "sess-1" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestVerifyNotImplementedWhenUnconfigured(t *testing.T) {
	t.Parallel()

	usecase := NewUsecase(nil, nil)
	_, err := usecase.Execute(context.Background(), "token", time.Now().UTC())
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestVerifyUnauthorizedCases(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	t.Run("empty token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)
		usecase := NewUsecase(authtest.Provider(), sessionmocks.NewMockSessionStore(ctrl))
		_, err := usecase.Execute(context.Background(), "   ", now)
		if !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("bad token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)
		usecase := NewUsecase(authtest.Provider(), sessionmocks.NewMockSessionStore(ctrl))
		_, err := usecase.Execute(context.Background(), "not-a-jwt", now)
		if !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("session not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)
		sessions := sessionmocks.NewMockSessionStore(ctrl)
		sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(nil, common.ErrNotFound)
		usecase := NewUsecase(authtest.Provider(), sessions)
		_, err := usecase.Execute(context.Background(), issueToken(t, "user-1", "sess-1", now), now)
		if !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("user mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)
		sessions := sessionmocks.NewMockSessionStore(ctrl)
		sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(&sessionmodel.Session{
			ID: "sess-1", UserID: "other", RefreshExpiresAt: now.Add(time.Hour),
		}, nil)
		usecase := NewUsecase(authtest.Provider(), sessions)
		_, err := usecase.Execute(context.Background(), issueToken(t, "user-1", "sess-1", now), now)
		if !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("expired session", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)
		sessions := sessionmocks.NewMockSessionStore(ctrl)
		sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(&sessionmodel.Session{
			ID: "sess-1", UserID: "user-1", RefreshExpiresAt: now.Add(-time.Minute),
		}, nil)
		usecase := NewUsecase(authtest.Provider(), sessions)
		_, err := usecase.Execute(context.Background(), issueToken(t, "user-1", "sess-1", now), now)
		if !errors.Is(err, common.ErrUnauthorized) {
			t.Fatalf("expected ErrUnauthorized, got %v", err)
		}
	})
}

func TestVerifyPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	now := time.Now().UTC()
	repoErr := errors.New("db down")
	sessions := sessionmocks.NewMockSessionStore(ctrl)
	sessions.EXPECT().Get(gomock.Any(), "sess-1").Return(nil, repoErr)

	usecase := NewUsecase(authtest.Provider(), sessions)
	_, err := usecase.Execute(context.Background(), issueToken(t, "user-1", "sess-1", now), now)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

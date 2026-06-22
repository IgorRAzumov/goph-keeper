package fake_test

import (
	"context"
	"testing"
	"time"

	"goph-keeper/internal/server/domain/common"
	sessionmodel "goph-keeper/internal/server/domain/session/model"
	"goph-keeper/internal/testutil/fake"
)

func TestSessionRepoRotateAndFind(t *testing.T) {
	t.Parallel()

	repo := fake.NewSessionRepo()
	now := time.Now().UTC()
	old := &sessionmodel.Session{
		ID: "s1", UserID: "u1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("old"),
		RefreshExpiresAt: now.Add(time.Hour),
	}
	_ = repo.Save(context.Background(), old)

	newSess := &sessionmodel.Session{
		ID: "s2", UserID: "u1",
		RefreshTokenHash: sessionmodel.HashRefreshToken("new"),
		RefreshExpiresAt: now.Add(2 * time.Hour),
	}
	if err := repo.Rotate(context.Background(), "s1", newSess); err != nil {
		t.Fatal(err)
	}
	found, err := repo.FindActiveByRefreshToken(context.Background(), "new", now)
	if err != nil || found.ID != "s2" {
		t.Fatalf("find: %+v err=%v", found, err)
	}
	if _, err := repo.Get(context.Background(), "s1"); err != common.ErrNotFound {
		t.Fatalf("expected old deleted, got %v", err)
	}
}

func TestSessionRepoDeleteNotFound(t *testing.T) {
	t.Parallel()

	if err := fake.NewSessionRepo().Delete(context.Background(), "missing"); err != common.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

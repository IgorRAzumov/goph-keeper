package fake_test

import (
	"context"
	"testing"

	"goph-keeper/internal/server/domain/common"
	recordmodel "goph-keeper/internal/server/domain/record/model"
	usermodel "goph-keeper/internal/server/domain/user/model"
	"goph-keeper/internal/testutil/fake"

	"golang.org/x/crypto/bcrypt"
)

func TestUserRepoNotFound(t *testing.T) {
	t.Parallel()

	_, err := fake.NewUserRepo().GetByLogin(context.Background(), "missing")
	if err != common.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestRecordRepoGetNotFound(t *testing.T) {
	t.Parallel()

	_, err := fake.NewRecordRepo().Get(context.Background(), "o", "r")
	if err != common.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestUserRepoSaveAndGet(t *testing.T) {
	t.Parallel()

	repo := fake.NewUserRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("p"), bcrypt.MinCost)
	err := repo.Save(context.Background(), &usermodel.User{ID: "u1", Login: "alice", PasswordHash: hash})
	if err != nil {
		t.Fatal(err)
	}
	u, err := repo.GetByLogin(context.Background(), "alice")
	if err != nil || u.ID != "u1" {
		t.Fatalf("get: %+v err=%v", u, err)
	}
}

func TestUserRepoConflict(t *testing.T) {
	t.Parallel()

	repo := fake.NewUserRepo()
	_ = repo.Save(context.Background(), &usermodel.User{ID: "u1", Login: "dup", PasswordHash: []byte("h")})
	err := repo.Save(context.Background(), &usermodel.User{ID: "u2", Login: "dup", PasswordHash: []byte("h")})
	if err != common.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestRecordRepoListSinceAndConflict(t *testing.T) {
	t.Parallel()

	repo := fake.NewRecordRepo()
	owner := "o1"
	_ = repo.Update(context.Background(), owner, &recordmodel.Record{
		ID: "r1", Type: recordmodel.RecordTypeText, Ciphertext: []byte("a"), Version: 5,
	})
	list, err := repo.ListSince(context.Background(), owner, 3)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %+v err=%v", list, err)
	}
	conflicts, err := repo.UpdateBatch(context.Background(), owner, []*recordmodel.Record{{
		ID: "r1", Type: recordmodel.RecordTypeText, Ciphertext: []byte("b"), Version: 3,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) != 1 || conflicts[0].Version != 5 {
		t.Fatalf("expected conflict with v5, got %+v", conflicts)
	}
}

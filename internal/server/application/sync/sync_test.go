package sync

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/record/model"
	repomocks "goph-keeper/internal/server/domain/record/repository/mocks"
	recordsvc "goph-keeper/internal/server/domain/record/service"

	"go.uber.org/mock/gomock"
)

func TestSyncPull(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	want := []*model.Record{{ID: "r1", OwnerID: "owner-1", Version: 2}}
	repo := repomocks.NewMockRecordStore(ctrl)
	repo.EXPECT().ListSince(gomock.Any(), "owner-1", int64(1)).Return(want, nil)

	usecase := NewUsecase(recordsvc.NewRecordService(repo))
	out, err := usecase.Pull(context.Background(), PullInput{OwnerID: "owner-1", SinceVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Records) != 1 || out.Records[0].ID != "r1" {
		t.Fatalf("unexpected pull output: %+v", out)
	}
}

func TestSyncPushReturnsConflict(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	serverRecord := &model.Record{
		ID: "r1", OwnerID: "owner-1", Type: model.RecordTypeText,
		Ciphertext: []byte("server"), Version: 5,
	}
	clientRecord := &model.Record{
		ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("client"), Version: 3,
	}

	repo := repomocks.NewMockRecordStore(ctrl)
	repo.EXPECT().UpdateBatch(gomock.Any(), "owner-1", gomock.Any()).Return([]*model.Record{serverRecord}, nil)

	usecase := NewUsecase(recordsvc.NewRecordService(repo))
	out, err := usecase.Push(context.Background(), PushInput{OwnerID: "owner-1", Records: []*model.Record{clientRecord}})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Conflicts) != 1 || out.Conflicts[0].Version != 5 {
		t.Fatalf("unexpected conflicts: %+v", out.Conflicts)
	}
}

func TestSyncPushSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	clientRecord := &model.Record{
		ID: "r1", Type: model.RecordTypeText, Ciphertext: []byte("client"), Version: 3,
	}

	repo := repomocks.NewMockRecordStore(ctrl)
	repo.EXPECT().UpdateBatch(gomock.Any(), "owner-1", gomock.Any()).Return([]*model.Record{}, nil)

	usecase := NewUsecase(recordsvc.NewRecordService(repo))
	out, err := usecase.Push(context.Background(), PushInput{OwnerID: "owner-1", Records: []*model.Record{clientRecord}})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %+v", out.Conflicts)
	}
}

func TestSyncPullRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	usecase := NewUsecase(recordsvc.NewRecordService(nil))
	_, err := usecase.Pull(context.Background(), PullInput{SinceVersion: 0})
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestSyncPullNotImplemented(t *testing.T) {
	t.Parallel()

	usecase := NewUsecase(nil)
	_, err := usecase.Pull(context.Background(), PullInput{OwnerID: "owner-1", SinceVersion: 0})
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestSyncPushNotImplemented(t *testing.T) {
	t.Parallel()

	usecase := NewUsecase(nil)
	_, err := usecase.Push(context.Background(), PushInput{OwnerID: "owner-1"})
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestSyncPushRejectsEmptyOwner(t *testing.T) {
	t.Parallel()

	usecase := NewUsecase(recordsvc.NewRecordService(nil))
	_, err := usecase.Push(context.Background(), PushInput{})
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

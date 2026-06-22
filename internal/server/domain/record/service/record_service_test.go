package service

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/record/model"
	repomocks "goph-keeper/internal/server/domain/record/repository/mocks"

	"go.uber.org/mock/gomock"
)

func TestGetReturnsRecord(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	want := &model.Record{ID: "record-1", OwnerID: "owner-1", Type: model.RecordTypeText}
	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().
		Get(gomock.Any(), "owner-1", "record-1").
		Return(want, nil)

	service := NewRecordService(repo)

	got, err := service.Get(context.Background(), "owner-1", "record-1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got != want {
		t.Fatalf("expected record pointer %p, got %p", want, got)
	}
}

func TestGetRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	service := NewRecordService(repomocks.NewMockRecordRepository(ctrl))

	tests := []struct {
		name     string
		ownerID  string
		recordID string
	}{
		{name: "empty owner", recordID: "record-1"},
		{name: "empty record", ownerID: "owner-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Get(context.Background(), tt.ownerID, tt.recordID)
			if !errors.Is(err, common.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestGetReturnsNotImplementedWithoutRepository(t *testing.T) {
	t.Parallel()

	service := NewRecordService(nil)

	_, err := service.Get(context.Background(), "owner-1", "record-1")
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestUpdateAssignsOwnerAndValidatesRecord(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().
		Update(gomock.Any(), "owner-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, ownerID string, record *model.Record) error {
			if ownerID != "owner-1" {
				t.Fatalf("unexpected owner id: %q", ownerID)
			}
			if record.OwnerID != "owner-1" {
				t.Fatalf("expected owner to be assigned, got %q", record.OwnerID)
			}
			return nil
		})

	service := NewRecordService(repo)

	err := service.Update(context.Background(), "owner-1", &model.Record{
		ID:         "record-1",
		Type:       model.RecordTypeText,
		Ciphertext: []byte("encrypted"),
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
}

func TestUpdateRejectsOwnerMismatch(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	service := NewRecordService(repo)

	err := service.Update(context.Background(), "owner-1", &model.Record{
		ID:         "record-1",
		OwnerID:    "owner-2",
		Type:       model.RecordTypeText,
		Ciphertext: []byte("encrypted"),
	})
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateRejectsUnknownType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	service := NewRecordService(repo)

	err := service.Update(context.Background(), "owner-1", &model.Record{
		ID:         "record-1",
		Type:       model.RecordType("unknown"),
		Ciphertext: []byte("encrypted"),
	})
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateRejectsEmptyCiphertextForActiveRecord(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	service := NewRecordService(repo)

	err := service.Update(context.Background(), "owner-1", &model.Record{
		ID:   "record-1",
		Type: model.RecordTypeText,
	})
	if !errors.Is(err, common.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateAllowsDeletedRecordWithoutCiphertext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().
		Update(gomock.Any(), "owner-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, record *model.Record) error {
			if !record.Deleted {
				t.Fatal("expected deleted record")
			}
			return nil
		})

	service := NewRecordService(repo)

	err := service.Update(context.Background(), "owner-1", &model.Record{
		ID:      "record-1",
		Type:    model.RecordTypeText,
		Deleted: true,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
}

func TestUpdateReturnsNotImplementedWithoutRepository(t *testing.T) {
	t.Parallel()

	service := NewRecordService(nil)

	err := service.Update(context.Background(), "owner-1", &model.Record{
		ID:         "record-1",
		Type:       model.RecordTypeText,
		Ciphertext: []byte("encrypted"),
	})
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestListSinceReturnsRecords(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	want := []*model.Record{{ID: "record-1", OwnerID: "owner-1", Version: 3}}
	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().
		ListSince(gomock.Any(), "owner-1", int64(2)).
		Return(want, nil)

	service := NewRecordService(repo)
	got, err := service.ListSince(context.Background(), "owner-1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "record-1" {
		t.Fatalf("unexpected records: %+v", got)
	}
}

func TestListSinceRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	service := NewRecordService(repomocks.NewMockRecordRepository(ctrl))

	tests := []struct {
		name         string
		ownerID      string
		sinceVersion int64
	}{
		{name: "empty owner", sinceVersion: 0},
		{name: "negative since", ownerID: "owner-1", sinceVersion: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ListSince(context.Background(), tt.ownerID, tt.sinceVersion)
			if !errors.Is(err, common.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestListSinceReturnsNotImplementedWithoutRepository(t *testing.T) {
	t.Parallel()

	service := NewRecordService(nil)
	_, err := service.ListSince(context.Background(), "owner-1", 0)
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestUpdateBatchReturnsConflicts(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	conflict := &model.Record{ID: "record-1", OwnerID: "owner-1", Version: 5}
	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().
		UpdateBatch(gomock.Any(), "owner-1", gomock.Any()).
		Return([]*model.Record{conflict}, nil)

	service := NewRecordService(repo)
	got, err := service.UpdateBatch(context.Background(), "owner-1", []*model.Record{{
		ID:         "record-1",
		Type:       model.RecordTypeText,
		Ciphertext: []byte("encrypted"),
		Version:    3,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Version != 5 {
		t.Fatalf("unexpected conflicts: %+v", got)
	}
}

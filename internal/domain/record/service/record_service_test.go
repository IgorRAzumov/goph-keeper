package service

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/record/model"
)

type recordRepoStub struct {
	update func(context.Context, string, *model.Record) error
	get    func(context.Context, string, string) (*model.Record, error)
}

func (r recordRepoStub) Update(ctx context.Context, ownerID string, record *model.Record) error {
	return r.update(ctx, ownerID, record)
}

func (r recordRepoStub) Get(ctx context.Context, ownerID, recordID string) (*model.Record, error) {
	return r.get(ctx, ownerID, recordID)
}

func TestGetReturnsRecord(t *testing.T) {
	t.Parallel()

	want := &model.Record{ID: "record-1", OwnerID: "owner-1", Type: model.RecordTypeText}
	service := NewRecordService(recordRepoStub{
		get: func(_ context.Context, ownerID, recordID string) (*model.Record, error) {
			if ownerID != "owner-1" || recordID != "record-1" {
				t.Fatalf("unexpected ids: owner=%q record=%q", ownerID, recordID)
			}
			return want, nil
		},
	})

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

	service := NewRecordService(recordRepoStub{})

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

	service := NewRecordService(recordRepoStub{
		update: func(_ context.Context, ownerID string, record *model.Record) error {
			if ownerID != "owner-1" {
				t.Fatalf("unexpected owner id: %q", ownerID)
			}
			if record.OwnerID != "owner-1" {
				t.Fatalf("expected owner to be assigned, got %q", record.OwnerID)
			}
			return nil
		},
	})

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

	service := NewRecordService(recordRepoStub{
		update: func(context.Context, string, *model.Record) error {
			t.Fatal("update must not be called for invalid input")
			return nil
		},
	})

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

	service := NewRecordService(recordRepoStub{
		update: func(context.Context, string, *model.Record) error {
			t.Fatal("update must not be called for invalid input")
			return nil
		},
	})

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

	service := NewRecordService(recordRepoStub{
		update: func(context.Context, string, *model.Record) error {
			t.Fatal("update must not be called for invalid input")
			return nil
		},
	})

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

	service := NewRecordService(recordRepoStub{
		update: func(_ context.Context, _ string, record *model.Record) error {
			if !record.Deleted {
				t.Fatal("expected deleted record")
			}
			return nil
		},
	})

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

package record

import (
	"context"
	"errors"
	"testing"

	"goph-keeper/internal/domain/common"
	"goph-keeper/internal/domain/record/model"
	repomocks "goph-keeper/internal/domain/record/repository/mocks"
	recordsvc "goph-keeper/internal/domain/record/service"

	"go.uber.org/mock/gomock"
)

func TestRecordUseCaseRead(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	want := &model.Record{ID: "record-1", OwnerID: "owner-1", Type: model.RecordTypeText}
	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().
		Get(gomock.Any(), "owner-1", "record-1").
		Return(want, nil)

	usecase := NewRecordUseCase(recordsvc.NewRecordService(repo))

	out, err := usecase.Read(context.Background(), GetRecordInput{OwnerID: "owner-1", RecordID: "record-1"})
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if out.Record != want {
		t.Fatalf("expected record pointer %p, got %p", want, out.Record)
	}
}

func TestRecordUseCaseReadRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	usecase := NewRecordUseCase(recordsvc.NewRecordService(nil))

	tests := []struct {
		name string
		in   GetRecordInput
	}{
		{name: "empty owner", in: GetRecordInput{RecordID: "record-1"}},
		{name: "empty record", in: GetRecordInput{OwnerID: "owner-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := usecase.Read(context.Background(), tt.in)
			if !errors.Is(err, common.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRecordUseCaseReadReturnsNotImplementedWithoutService(t *testing.T) {
	t.Parallel()

	usecase := NewRecordUseCase(nil)

	_, err := usecase.Read(context.Background(), GetRecordInput{OwnerID: "owner-1", RecordID: "record-1"})
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestRecordUseCaseUpdateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	usecase := NewRecordUseCase(recordsvc.NewRecordService(nil))

	tests := []struct {
		name string
		in   UpdateRecordInput
	}{
		{name: "empty owner", in: UpdateRecordInput{Record: &model.Record{ID: "record-1"}}},
		{name: "nil record", in: UpdateRecordInput{OwnerID: "owner-1"}},
		{name: "empty record id", in: UpdateRecordInput{OwnerID: "owner-1", Record: &model.Record{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := usecase.Update(context.Background(), tt.in)
			if !errors.Is(err, common.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestRecordUseCaseUpdate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().
		Update(gomock.Any(), "owner-1", gomock.Any()).
		DoAndReturn(func(_ context.Context, ownerID string, record *model.Record) error {
			if ownerID != "owner-1" {
				t.Fatalf("unexpected owner: %q", ownerID)
			}
			if record.ID != "record-1" {
				t.Fatalf("unexpected record id: %q", record.ID)
			}
			return nil
		})

	usecase := NewRecordUseCase(recordsvc.NewRecordService(repo))

	_, err := usecase.Update(context.Background(), UpdateRecordInput{
		OwnerID: "owner-1",
		Record: &model.Record{
			ID:         "record-1",
			Type:       model.RecordTypeText,
			Ciphertext: []byte("encrypted"),
		},
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
}

func TestRecordUseCaseUpdateReturnsNotImplementedWithoutService(t *testing.T) {
	t.Parallel()

	usecase := NewRecordUseCase(nil)

	_, err := usecase.Update(context.Background(), UpdateRecordInput{
		OwnerID: "owner-1",
		Record:  &model.Record{ID: "record-1"},
	})
	if !errors.Is(err, common.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

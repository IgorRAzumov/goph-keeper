package mocks

import (
	"context"
	"testing"

	recordmodel "goph-keeper/internal/server/domain/record/model"

	"go.uber.org/mock/gomock"
)

func TestMockRecordStoreMethods(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mock := NewMockRecordStore(ctrl)
	if mock.EXPECT() == nil {
		t.Fatal("expected recorder")
	}

	record := &recordmodel.Record{ID: "r1", OwnerID: "u1", Type: recordmodel.RecordTypeText, Version: 1}
	mock.EXPECT().Get(gomock.Any(), "u1", "r1").Return(record, nil)
	mock.EXPECT().ListSince(gomock.Any(), "u1", int64(0)).Return([]*recordmodel.Record{record}, nil)
	mock.EXPECT().Update(gomock.Any(), "u1", gomock.Any()).Return(nil)
	mock.EXPECT().UpdateBatch(gomock.Any(), "u1", gomock.Any()).Return([]*recordmodel.Record{}, nil)

	if _, err := mock.Get(context.Background(), "u1", "r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.ListSince(context.Background(), "u1", 0); err != nil {
		t.Fatal(err)
	}
	if err := mock.Update(context.Background(), "u1", record); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.UpdateBatch(context.Background(), "u1", []*recordmodel.Record{record}); err != nil {
		t.Fatal(err)
	}
}

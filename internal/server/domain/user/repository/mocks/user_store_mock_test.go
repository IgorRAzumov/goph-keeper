package mocks

import (
	"context"
	"testing"

	"goph-keeper/internal/server/domain/user/model"

	"go.uber.org/mock/gomock"
)

func TestMockUserStoreMethods(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mock := NewMockUserStore(ctrl)
	if mock.EXPECT() == nil {
		t.Fatal("expected recorder")
	}

	mock.EXPECT().GetByLogin(gomock.Any(), "alice").Return(&model.User{ID: "u1", Login: "alice"}, nil)
	mock.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	got, err := mock.GetByLogin(context.Background(), "alice")
	if err != nil || got == nil || got.ID != "u1" {
		t.Fatalf("unexpected GetByLogin result: user=%+v err=%v", got, err)
	}
	if err := mock.Save(context.Background(), &model.User{ID: "u2", Login: "bob"}); err != nil {
		t.Fatal(err)
	}
}

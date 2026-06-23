package mocks

import (
	"context"
	"testing"
	"time"

	sessionmodel "goph-keeper/internal/server/domain/session/model"

	"go.uber.org/mock/gomock"
)

func TestMockSessionStoreMethods(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mock := NewMockSessionStore(ctrl)
	if mock.EXPECT() == nil {
		t.Fatal("expected recorder")
	}

	now := time.Now().UTC()
	session := &sessionmodel.Session{ID: "s1", UserID: "u1", RefreshExpiresAt: now.Add(time.Hour)}

	mock.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	mock.EXPECT().Get(gomock.Any(), "s1").Return(session, nil)
	mock.EXPECT().Delete(gomock.Any(), "s1").Return(nil)
	mock.EXPECT().Rotate(gomock.Any(), "s1", gomock.Any()).Return(nil)
	mock.EXPECT().FindActiveByRefreshToken(gomock.Any(), "token", gomock.Any()).Return(session, nil)

	if err := mock.Save(context.Background(), session); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Get(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.Delete(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.Rotate(context.Background(), "s1", session); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.FindActiveByRefreshToken(context.Background(), "token", now); err != nil {
		t.Fatal(err)
	}
}

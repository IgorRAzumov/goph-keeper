package model

import "testing"

func TestUserModel(t *testing.T) {
	t.Parallel()

	u := User{ID: "u1", Login: "alice", PasswordHash: []byte("h")}
	if u.ID == "" || u.Login == "" {
		t.Fatal("unexpected empty user fields")
	}
}

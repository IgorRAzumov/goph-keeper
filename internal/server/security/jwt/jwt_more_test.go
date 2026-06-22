package jwt

import (
	"strings"
	"testing"
	"time"
)

func TestIssueAccessTokenValidation(t *testing.T) {
	t.Parallel()

	p := NewProvider("secret")
	now := time.Now().UTC()

	_, err := p.IssueAccessToken("", "s", time.Minute, now)
	if err == nil {
		t.Fatal("expected empty user id error")
	}
	_, err = p.IssueAccessToken("u", "", time.Minute, now)
	if err == nil {
		t.Fatal("expected empty session id error")
	}
	_, err = p.IssueAccessToken("u", "s", 0, now)
	if err == nil {
		t.Fatal("expected invalid ttl error")
	}
	_, err = NewProvider("").IssueAccessToken("u", "s", time.Minute, now)
	if err == nil {
		t.Fatal("expected empty secret error")
	}
}

func TestParseAccessTokenMalformed(t *testing.T) {
	t.Parallel()

	_, err := NewProvider("secret").ParseAccessToken("not.a.jwt", time.Now().UTC())
	if err == nil || !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("expected malformed error, got %v", err)
	}
}

package jwt

import (
	"strings"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestIssueAndParseAccessToken(t *testing.T) {
	t.Parallel()

	p := NewProvider("secret")
	now := time.Unix(1_700_000_000, 0).UTC()

	token, err := p.IssueAccessToken("user-1", "sess-1", time.Minute, now)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	claims, err := p.ParseAccessToken(token, now.Add(30*time.Second))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Sub != "user-1" || claims.Sid != "sess-1" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	if _, err := p.ParseAccessToken(token, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expected expired")
	}
}

func TestParseAccessTokenRejectsInvalidHeader(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0).UTC()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS384, accessClaims{
		Sub: "user-1",
		Sid: "sess-1",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(now),
		},
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := NewProvider("secret").ParseAccessToken(signed, now); err == nil {
		t.Fatal("expected invalid header error")
	}
}

func TestParseAccessTokenRejectsBadSignature(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	token, err := NewProvider("secret").IssueAccessToken("user-1", "sess-1", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewProvider("other-secret").ParseAccessToken(token, now)
	if err == nil {
		t.Fatal("expected bad signature error")
	}
	if !strings.Contains(err.Error(), "bad signature") {
		t.Fatalf("expected bad signature error, got %v", err)
	}
}

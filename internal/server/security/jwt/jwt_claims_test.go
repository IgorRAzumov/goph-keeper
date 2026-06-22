package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestParseAccessTokenMissingClaims(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, accessClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Minute)),
			IssuedAt:  jwtlib.NewNumericDate(now),
		},
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewProvider("secret").ParseAccessToken(signed, now)
	if err == nil {
		t.Fatal("expected missing claims error")
	}
}

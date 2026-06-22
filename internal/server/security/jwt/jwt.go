package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims — минимальный набор полей access JWT.
type Claims struct {
	Sub string `json:"sub"` // user id
	Sid string `json:"sid"` // session id
	Exp int64  `json:"exp"` // unix seconds
	Iat int64  `json:"iat"` // unix seconds
}

type accessClaims struct {
	Sub string `json:"sub"`
	Sid string `json:"sid"`
	jwtlib.RegisteredClaims
}

// Provider подписывает и проверяет access JWT (HS256) через github.com/golang-jwt/jwt/v5.
type Provider struct {
	secret []byte
}

func NewProvider(secret string) *Provider {
	return &Provider{secret: []byte(secret)}
}

// IssueAccessToken выпускает access JWT.
func (provider *Provider) IssueAccessToken(userID, sessionID string, ttl time.Duration, now time.Time) (string, error) {
	if provider == nil || len(provider.secret) == 0 {
		return "", errors.New("jwt: empty secret")
	}
	if userID == "" || sessionID == "" {
		return "", errors.New("jwt: empty claims")
	}
	if ttl <= 0 {
		return "", errors.New("jwt: invalid ttl")
	}

	claims := accessClaims{
		Sub: userID,
		Sid: sessionID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwtlib.NewNumericDate(now),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(provider.secret)
}

// ParseAccessToken проверяет подпись и парсит claims.
func (provider *Provider) ParseAccessToken(token string, now time.Time) (Claims, error) {
	if provider == nil || len(provider.secret) == 0 {
		return Claims{}, errors.New("jwt: empty secret")
	}

	var parsed accessClaims
	parser := jwtlib.NewParser(
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithExpirationRequired(),
		jwtlib.WithTimeFunc(func() time.Time { return now }),
	)

	_, err := parser.ParseWithClaims(token, &parsed, func(token *jwtlib.Token) (any, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, fmt.Errorf("jwt: invalid header")
		}
		return provider.secret, nil
	})
	if err != nil {
		return Claims{}, mapParseError(err)
	}

	if parsed.Sub == "" || parsed.Sid == "" {
		return Claims{}, errors.New("jwt: missing claims")
	}

	out := Claims{Sub: parsed.Sub, Sid: parsed.Sid}
	if parsed.ExpiresAt != nil {
		out.Exp = parsed.ExpiresAt.Unix()
	}
	if parsed.IssuedAt != nil {
		out.Iat = parsed.IssuedAt.Unix()
	}
	return out, nil
}

func mapParseError(err error) error {
	switch {
	case errors.Is(err, jwtlib.ErrTokenExpired):
		return errors.New("jwt: expired")
	case errors.Is(err, jwtlib.ErrTokenSignatureInvalid):
		return errors.New("jwt: bad signature")
	case errors.Is(err, jwtlib.ErrTokenMalformed):
		return errors.New("jwt: invalid token format")
	case errors.Is(err, jwtlib.ErrTokenUnverifiable):
		return errors.New("jwt: invalid header")
	default:
		return err
	}
}

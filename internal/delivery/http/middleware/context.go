package middleware

import "context"

type authCtxKey int

const (
	userIDKey authCtxKey = iota + 1
	sessionIDKey
)

// WithUserID кладёт user id в контекст запроса.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserID достаёт user id из контекста (пустая строка, если нет).
func UserID(ctx context.Context) string {
	v := ctx.Value(userIDKey)
	s, _ := v.(string)
	return s
}

// WithSessionID кладёт session id в контекст запроса.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// SessionID достаёт session id из контекста (пустая строка, если нет).
func SessionID(ctx context.Context) string {
	v := ctx.Value(sessionIDKey)
	s, _ := v.(string)
	return s
}

package auth

import "context"

type UserClaims struct {
	Subject string
}

type userContextKey struct{}

var userKey userContextKey

func WithClaims(ctx context.Context, u UserClaims) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func ClaimsFromContext(ctx context.Context) (UserClaims, bool) {
	user, ok := ctx.Value(userKey).(UserClaims)
	return user, ok
}

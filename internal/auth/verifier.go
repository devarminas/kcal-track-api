package auth

import "context"

type Verifier interface {
	VerifySession(ctx context.Context, token string) (UserClaims, error)
}

package contextutil

import (
	"context"

	"kseli-server/auth"
)

type ContextKey string

const UserClaimsKey ContextKey = "userClaims"

func GetUserClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims)
	return claims, ok
}

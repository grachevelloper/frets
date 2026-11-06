package login

import (
	"context"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type Provider interface {
	AuthorizationURL(state string, verifier string) (location string)
	Identity(ctx context.Context, code string, verifier string) (identity auth.Identity, err error)
}

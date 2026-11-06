package session

import (
	"context"
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type Repository interface {
	CreateSession(ctx context.Context, provider string, identity auth.Identity, tokenHash []byte, expiresAt time.Time) (err error)
	Profile(ctx context.Context, tokenHash []byte) (profile auth.Profile, err error)
	Revoke(ctx context.Context, tokenHash []byte) (err error)
}

type TokenGenerator interface {
	Generate() (token string)
}

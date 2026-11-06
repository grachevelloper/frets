package login

import (
	"context"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type Repository interface {
	SaveAttempt(ctx context.Context, attempt auth.Attempt) (err error)
	ConsumeAttempt(ctx context.Context, stateHash []byte, browserHash []byte, provider string) (attempt auth.Attempt, err error)
}

type Sessions interface {
	Create(ctx context.Context, provider string, identity auth.Identity) (sessionToken string, err error)
}

type TokenGenerator interface {
	Generate() (token string)
}

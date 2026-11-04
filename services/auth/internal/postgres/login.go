package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/services/auth/internal/account"
)

func (r *Repository) SaveAttempt(ctx context.Context, attempt account.Attempt) error {
	if _, err := r.db.Exec(ctx, "DELETE FROM auth.login_attempts WHERE expires_at <= now()"); err != nil {
		return err
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO auth.login_attempts(state_hash,provider,browser_hash,verifier,return_to,expires_at)
		VALUES($1,$2,$3,$4,$5,$6)
	`, attempt.StateHash, attempt.Provider, attempt.BrowserHash, attempt.Verifier, attempt.ReturnTo, attempt.ExpiresAt)
	return err
}

func (r *Repository) ConsumeAttempt(ctx context.Context, stateHash, browserHash []byte, provider string) (account.Attempt, error) {
	var attempt account.Attempt
	err := r.db.QueryRow(ctx, `
		DELETE FROM auth.login_attempts
		WHERE state_hash=$1 AND provider=$2 AND browser_hash=$3 AND expires_at>now()
		RETURNING verifier,return_to
	`, stateHash, provider, browserHash).Scan(&attempt.Verifier, &attempt.ReturnTo)
	if errors.Is(err, pgx.ErrNoRows) {
		return attempt, account.ErrLoginInvalid
	}
	return attempt, err
}

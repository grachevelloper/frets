package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

func (r *Repository) Profile(ctx context.Context, tokenHash []byte) (auth.Profile, error) {
	var p auth.Profile
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.nickname, u.email, u.created_at
		FROM auth.sessions s JOIN auth.users u ON u.id = s.user_id
		WHERE s.token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at > now()
	`, tokenHash).Scan(&p.ID, &p.Nickname, &p.Email, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, auth.ErrUnauthorized
	}
	return p, err
}

func (r *Repository) Revoke(ctx context.Context, tokenHash []byte) error {
	result, err := r.db.Exec(ctx, `
		UPDATE auth.sessions SET revoked_at=now()
		WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at>now()
	`, tokenHash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return auth.ErrUnauthorized
	}
	return nil
}

func (r *Repository) CreateSession(ctx context.Context, provider string, identity auth.Identity, tokenHash []byte, expiresAt time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Serialize first sign-in for this identity so concurrent callbacks create only one profile.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 1))", provider+":"+identity.ID); err != nil {
		return err
	}

	var userID string
	err = tx.QueryRow(ctx, "SELECT user_id FROM auth.identities WHERE provider=$1 AND provider_account_id=$2", provider, identity.ID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, "INSERT INTO auth.users(nickname,email) VALUES($1,$2) RETURNING id", identity.Name, identity.Email).Scan(&userID)
		if err == nil {
			_, err = tx.Exec(ctx, "INSERT INTO auth.identities VALUES($1,$2,$3)", provider, identity.ID, userID)
		}
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "INSERT INTO auth.sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)", tokenHash, userID, expiresAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

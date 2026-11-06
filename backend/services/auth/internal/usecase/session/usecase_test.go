package session

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type sessionRepository struct {
	key     []byte
	expires time.Time
}

func (r *sessionRepository) CreateSession(_ context.Context, provider string, identity auth.Identity, key []byte, expires time.Time) error {
	r.key = key
	r.expires = expires
	return nil
}

func (r *sessionRepository) Profile(_ context.Context, key []byte) (auth.Profile, error) {
	if !bytes.Equal(key, r.key) {
		return auth.Profile{}, auth.ErrUnauthorized
	}
	return auth.Profile{ID: "user"}, nil
}

func (r *sessionRepository) Revoke(_ context.Context, key []byte) error {
	if !bytes.Equal(key, r.key) {
		return auth.ErrUnauthorized
	}
	r.key = nil
	return nil
}

type tokenGenerator struct{}

func (tokenGenerator) Generate() string { return "raw-session-secret" }

func TestSessionUsesHashedTokensAndRevocation(t *testing.T) {
	repo := &sessionRepository{}
	u := New(repo, tokenGenerator{})
	key, err := u.Create(t.Context(), "yandex", auth.Identity{ID: "account"})
	if err != nil || !bytes.Equal(repo.key, auth.HashSecret(key)) || bytes.Equal(repo.key, []byte(key)) {
		t.Fatalf("persisted token: key=%q error=%v", key, err)
	}
	if time.Until(repo.expires) < auth.SessionLifetime-time.Minute {
		t.Fatal("unexpected session lifetime")
	}
	if p, err := u.Profile(t.Context(), key); err != nil || p.ID != "user" {
		t.Fatalf("profile=%+v error=%v", p, err)
	}
	if err := u.Revoke(t.Context(), key); err != nil {
		t.Fatal(err)
	}
	if _, err := u.Profile(t.Context(), key); !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatalf("revoked token: %v", err)
	}
}

func TestEmptyTokenRejectedWithoutRepository(t *testing.T) {
	u := New(nil, nil)
	if _, err := u.Profile(t.Context(), ""); !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatal(err)
	}
	if err := u.Revoke(t.Context(), ""); !errors.Is(err, auth.ErrUnauthorized) {
		t.Fatal(err)
	}
}

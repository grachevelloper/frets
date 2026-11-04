package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grachevelloper/frets/services/auth/internal/account"
)

func testRepository(t *testing.T) *Repository {
	t.Helper()
	dsn := os.Getenv("FRETS_TEST_AUTH_DATABASE_URL")
	if dsn == "" {
		t.Skip("FRETS_TEST_AUTH_DATABASE_URL is unset")
	}

	db, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return New(db)
}

func TestIdentityAndSessionPersistence(t *testing.T) {
	repo := testRepository(t)
	ctx := t.Context()
	email := uuid.NewString() + "@example.test"
	identity := account.Identity{ID: uuid.NewString(), Name: "Guitarist", Email: &email}
	t.Cleanup(func() {
		if _, err := repo.db.Exec(context.Background(), "DELETE FROM auth.sessions WHERE user_id IN (SELECT id FROM auth.users WHERE email=$1)", email); err != nil {
			t.Error(err)
		}
		if _, err := repo.db.Exec(context.Background(), "DELETE FROM auth.identities WHERE provider_account_id=$1", identity.ID); err != nil {
			t.Error(err)
		}
		if _, err := repo.db.Exec(context.Background(), "DELETE FROM auth.users WHERE email=$1", email); err != nil {
			t.Error(err)
		}
	})

	create := func(provider string, expires time.Time) []byte {
		t.Helper()
		key := []byte(uuid.NewString())
		if err := repo.CreateSession(ctx, provider, identity, key, expires); err != nil {
			t.Fatal(err)
		}
		return key
	}
	profile := func(key []byte) account.Profile {
		t.Helper()
		p, err := repo.Profile(ctx, key)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}

	expires := time.Now().Add(time.Hour)
	first := create("google", expires)
	second := create("google", expires)
	other := create("yandex", expires)
	if profile(first).ID != profile(second).ID || profile(first).ID == profile(other).ID {
		t.Fatal("profiles must follow provider account identity")
	}
	if err := repo.Revoke(ctx, first); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Profile(ctx, first); !errors.Is(err, account.ErrUnauthorized) {
		t.Fatalf("revoked session: %v", err)
	}
	expired := create("google", time.Now().Add(-time.Hour))
	if _, err := repo.Profile(ctx, expired); !errors.Is(err, account.ErrUnauthorized) {
		t.Fatalf("expired session: %v", err)
	}
}

func TestAttemptCanOnlyBeConsumedOnce(t *testing.T) {
	repo := testRepository(t)
	ctx := t.Context()
	a := account.Attempt{
		StateHash: []byte(uuid.NewString()), BrowserHash: []byte(uuid.NewString()), Provider: "yandex",
		Verifier: "verifier", ReturnTo: "/songs/example", ExpiresAt: time.Now().Add(time.Minute),
	}
	if err := repo.SaveAttempt(ctx, a); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := repo.db.Exec(context.Background(), "DELETE FROM auth.login_attempts WHERE state_hash=$1", a.StateHash); err != nil {
			t.Error(err)
		}
	})

	if _, err := repo.ConsumeAttempt(ctx, a.StateHash, []byte("wrong"), a.Provider); !errors.Is(err, account.ErrLoginInvalid) {
		t.Fatalf("foreign browser: %v", err)
	}
	result, err := repo.ConsumeAttempt(ctx, a.StateHash, a.BrowserHash, a.Provider)
	if err != nil || result.Verifier != a.Verifier || result.ReturnTo != a.ReturnTo {
		t.Fatalf("attempt = %+v, err = %v", result, err)
	}
	if _, err := repo.ConsumeAttempt(ctx, a.StateHash, a.BrowserHash, a.Provider); !errors.Is(err, account.ErrLoginInvalid) {
		t.Fatalf("replayed attempt: %v", err)
	}
}

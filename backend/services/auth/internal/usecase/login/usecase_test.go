package login

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type attemptRepository struct {
	attempt *auth.Attempt
}

func (r *attemptRepository) SaveAttempt(_ context.Context, attempt auth.Attempt) error {
	r.attempt = &attempt
	return nil
}

func (r *attemptRepository) ConsumeAttempt(_ context.Context, state, browser []byte, provider string) (auth.Attempt, error) {
	a := r.attempt
	if a == nil || !bytes.Equal(a.StateHash, state) || !bytes.Equal(a.BrowserHash, browser) || a.Provider != provider || time.Now().After(a.ExpiresAt) {
		return auth.Attempt{}, auth.ErrLoginInvalid
	}
	r.attempt = nil
	return *a, nil
}

type providerStub struct {
	state    string
	verifier string
	err      error
}

func (p *providerStub) AuthorizationURL(state, verifier string) string {
	p.state = state
	p.verifier = verifier
	return "https://identity.example"
}

func (p *providerStub) Identity(ctx context.Context, code, verifier string) (auth.Identity, error) {
	if code != "code" || verifier != p.verifier {
		return auth.Identity{}, errors.New("unexpected PKCE verifier")
	}
	return auth.Identity{ID: "provider-id"}, p.err
}

type sessionCreator struct {
	identity auth.Identity
}

func (s *sessionCreator) Create(_ context.Context, provider string, identity auth.Identity) (string, error) {
	s.identity = identity
	return "session-token", nil
}

type sequenceTokens struct {
	next int
}

func (g *sequenceTokens) Generate() string {
	g.next++
	return fmt.Sprintf("%043d", g.next)
}

func TestLoginBindsBrowserAndConsumesStateOnce(t *testing.T) {
	attempts := &attemptRepository{}
	provider := &providerStub{}
	sessions := &sessionCreator{}
	u := New(attempts, map[string]Provider{
		"yandex": provider,
		"google": nil,
	}, sessions, &sequenceTokens{})
	ctx := t.Context()
	if _, _, err := u.Begin(ctx, "google", ""); !errors.Is(err, auth.ErrProviderUnavailable) {
		t.Fatalf("unconfigured provider: %v", err)
	}
	const next = "/songs/song/fragments/chorus"
	_, browser, err := u.Begin(ctx, "yandex", next)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(attempts.attempt.StateHash, []byte(provider.state)) || bytes.Equal(attempts.attempt.BrowserHash, []byte(browser)) {
		t.Fatal("stored raw login secrets")
	}
	if _, _, consumed, err := u.Complete(ctx, "yandex", "other-browser", provider.state, "code"); consumed || !errors.Is(err, auth.ErrLoginInvalid) {
		t.Fatalf("foreign browser: consumed=%v error=%v", consumed, err)
	}
	key, path, consumed, err := u.Complete(ctx, "yandex", browser, provider.state, "code")
	if err != nil || !consumed || key != "session-token" || path != next || sessions.identity.ID != "provider-id" {
		t.Fatalf("callback: key=%q next=%q consumed=%v error=%v", key, path, consumed, err)
	}
	if _, _, _, err := u.Complete(ctx, "yandex", browser, provider.state, "code"); !errors.Is(err, auth.ErrLoginInvalid) {
		t.Fatalf("replayed callback: %v", err)
	}
}

func TestFailedProviderStillConsumesAttempt(t *testing.T) {
	attempts := &attemptRepository{}
	providerErr := errors.New("provider unavailable")
	provider := &providerStub{err: providerErr}
	u := New(attempts, map[string]Provider{"yandex": provider}, &sessionCreator{}, &sequenceTokens{})
	_, browser, err := u.Begin(t.Context(), "yandex", "/")
	if err != nil {
		t.Fatal(err)
	}
	_, _, consumed, err := u.Complete(t.Context(), "yandex", browser, provider.state, "code")
	if !consumed || !errors.Is(err, auth.ErrLoginFailed) || !errors.Is(err, providerErr) || attempts.attempt != nil {
		t.Fatalf("consumed=%v error=%v", consumed, err)
	}
}

func TestReturnPath(t *testing.T) {
	externalPaths := []string{"https://elsewhere.test/songs/1", "//elsewhere.test", "/%2felsewhere.test", "/songs/\\elsewhere", "/api/v1/me"}
	for _, path := range externalPaths {
		if got := returnPath(path); got != "/" {
			t.Errorf("returnPath(%q)=%q", path, got)
		}
	}
}

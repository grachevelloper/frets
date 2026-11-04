package account

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

type memoryAttempts struct {
	attempt *Attempt
}

func (m *memoryAttempts) SaveAttempt(_ context.Context, a Attempt) error {
	m.attempt = &a
	return nil
}

func (m *memoryAttempts) ConsumeAttempt(_ context.Context, state, browser []byte, provider string) (Attempt, error) {
	a := m.attempt
	if a == nil || !bytes.Equal(state, a.StateHash) || !bytes.Equal(browser, a.BrowserHash) || a.Provider != provider {
		return Attempt{}, ErrLoginInvalid
	}
	m.attempt = nil
	return *a, nil
}

type fakeProvider struct {
	t               *testing.T
	state, verifier string
}

func (p *fakeProvider) AuthorizationURL(state, verifier string) string {
	p.state, p.verifier = state, verifier
	return "https://identity.example"
}

func (p *fakeProvider) Identity(_ context.Context, code, verifier string) (Identity, error) {
	p.t.Helper()
	if code != "code" || verifier != p.verifier {
		p.t.Fatalf("unexpected code or PKCE verifier")
	}
	return Identity{ID: "provider-id", Name: "Profile"}, nil
}

type memorySessions struct {
	key      []byte
	provider string
	identity Identity
	expires  time.Time
}

func (m *memorySessions) CreateSession(_ context.Context, provider string, identity Identity, key []byte, expires time.Time) error {
	m.key, m.provider, m.identity, m.expires = key, provider, identity, expires
	return nil
}

func (m *memorySessions) Profile(_ context.Context, key []byte) (Profile, error) {
	if !bytes.Equal(key, m.key) {
		return Profile{}, ErrUnauthorized
	}
	return Profile{ID: "user"}, nil
}

func (m *memorySessions) Revoke(_ context.Context, key []byte) error {
	if !bytes.Equal(key, m.key) {
		return ErrUnauthorized
	}
	m.key = nil
	return nil
}

func TestLoginBindsStateToBrowserAndConsumesItOnce(t *testing.T) {
	ctx := context.Background()
	attempts := &memoryAttempts{}
	repo := &memorySessions{}
	sessions := NewSessions(repo)
	provider := &fakeProvider{t: t}
	login := NewLogin(attempts, map[string]Provider{"yandex": provider, "google": nil}, sessions)
	const nextPath = "/songs/song/fragments/chorus"

	if _, _, err := login.Begin(ctx, "google", ""); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("unconfigured provider: %v", err)
	}
	location, browser, err := login.Begin(ctx, "yandex", nextPath)
	if err != nil || location != "https://identity.example" || browser == "" || provider.state == "" {
		t.Fatalf("begin: location = %q, err = %v", location, err)
	}
	if bytes.Equal(attempts.attempt.StateHash, []byte(provider.state)) || bytes.Equal(attempts.attempt.BrowserHash, []byte(browser)) {
		t.Fatal("stored raw state or browser token")
	}

	if _, _, consumed, err := login.Complete(ctx, "yandex", "other-browser", provider.state, "code"); consumed || !errors.Is(err, ErrLoginInvalid) {
		t.Fatalf("cross-browser callback: consumed = %v, err = %v", consumed, err)
	}
	key, next, consumed, err := login.Complete(ctx, "yandex", browser, provider.state, "code")
	if err != nil || !consumed || next != nextPath || !bytes.Equal(repo.key, hash(key)) {
		t.Fatalf("complete: consumed = %v, next = %q, err = %v", consumed, next, err)
	}
	if repo.provider != "yandex" || repo.identity.ID != "provider-id" || time.Until(repo.expires) < SessionLifetime-time.Minute {
		t.Fatalf("unexpected persisted session")
	}
	if _, _, _, err := login.Complete(ctx, "yandex", browser, provider.state, "code"); !errors.Is(err, ErrLoginInvalid) {
		t.Fatalf("replayed callback: %v", err)
	}
	if profile, err := sessions.Profile(ctx, key); err != nil || profile.ID != "user" {
		t.Fatalf("profile = %+v, err = %v", profile, err)
	}
	if err := sessions.Revoke(ctx, key); err != nil {
		t.Fatal(err)
	}
	if _, err := sessions.Profile(ctx, key); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("revoked session: %v", err)
	}
}

func TestReturnPathRejectsExternalRedirects(t *testing.T) {
	for _, path := range []string{"https://elsewhere.test/songs/1", "//elsewhere.test", "/%2felsewhere.test", "/songs/\\elsewhere", "/api/v1/me"} {
		if got := returnPath(path); got != "/" {
			t.Errorf("returnPath(%q) = %q", path, got)
		}
	}
}

package login

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type UseCase struct {
	attempts  Repository
	providers map[string]Provider
	sessions  Sessions
	tokens    TokenGenerator
}

func New(attempts Repository, providers map[string]Provider, sessions Sessions, tokens TokenGenerator) *UseCase {
	return &UseCase{
		attempts:  attempts,
		providers: providers,
		sessions:  sessions,
		tokens:    tokens,
	}
}

func (s *UseCase) Begin(ctx context.Context, provider, next string) (location, browser string, err error) {
	oauth, known := s.providers[provider]
	if !known {
		return "", "", auth.ErrUnknownProvider
	}
	if oauth == nil {
		return "", "", auth.ErrProviderUnavailable
	}

	state, browser, verifier := s.tokens.Generate(), s.tokens.Generate(), s.tokens.Generate()
	attempt := auth.Attempt{
		StateHash:   auth.HashSecret(state),
		BrowserHash: auth.HashSecret(browser),
		Provider:    provider,
		Verifier:    verifier,
		ReturnTo:    returnPath(next),
		ExpiresAt:   time.Now().Add(auth.LoginLifetime),
	}
	if err := s.attempts.SaveAttempt(ctx, attempt); err != nil {
		return "", "", fmt.Errorf("save login attempt: %w", err)
	}
	return oauth.AuthorizationURL(state, verifier), browser, nil
}

func (s *UseCase) Complete(ctx context.Context, provider, browser, state, code string) (key, next string, consumed bool, err error) {
	oauth := s.providers[provider]
	if oauth == nil || browser == "" || state == "" {
		return "", "", false, auth.ErrLoginInvalid
	}

	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	attempt, err := s.attempts.ConsumeAttempt(ctx, auth.HashSecret(state), auth.HashSecret(browser), provider)
	if err != nil {
		return "", "", false, fmt.Errorf("consume login attempt: %w", err)
	}
	if code == "" {
		return "", "", true, auth.ErrLoginInvalid
	}

	identity, err := oauth.Identity(ctx, code, attempt.Verifier)
	if err != nil {
		return "", "", true, fmt.Errorf("exchange identity with %s: %w: %w", provider, auth.ErrLoginFailed, err)
	}
	key, err = s.sessions.Create(ctx, provider, identity)
	return key, attempt.ReturnTo, true, err
}

func returnPath(raw string) string {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host != "" || strings.Contains(parsed.Path, "\\") || strings.HasPrefix(parsed.Path, "//") {
		return "/"
	}
	if strings.HasPrefix(parsed.Path, "/songs/") {
		return parsed.EscapedPath()
	}
	return "/"
}

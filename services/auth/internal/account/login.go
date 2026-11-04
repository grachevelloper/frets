package account

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"
)

const LoginLifetime = 10 * time.Minute

var (
	ErrUnknownProvider     = errors.New("not_found")
	ErrProviderUnavailable = errors.New("provider_unavailable")
	ErrLoginInvalid        = errors.New("login_invalid")
	ErrLoginFailed         = errors.New("login_failed")
)

type Attempt struct {
	StateHash, BrowserHash       []byte
	Provider, Verifier, ReturnTo string
	ExpiresAt                    time.Time
}

type LoginAttempts interface {
	SaveAttempt(context.Context, Attempt) error
	ConsumeAttempt(context.Context, []byte, []byte, string) (Attempt, error)
}

type Provider interface {
	AuthorizationURL(state, verifier string) string
	Identity(ctx context.Context, code, verifier string) (Identity, error)
}

type LoginService struct {
	attempts  LoginAttempts
	providers map[string]Provider
	sessions  *SessionService
}

func NewLogin(attempts LoginAttempts, providers map[string]Provider, sessions *SessionService) *LoginService {
	return &LoginService{attempts: attempts, providers: providers, sessions: sessions}
}

func (s *LoginService) Begin(ctx context.Context, provider, next string) (location, browser string, err error) {
	oauth, known := s.providers[provider]
	if !known {
		return "", "", ErrUnknownProvider
	}
	if oauth == nil {
		return "", "", ErrProviderUnavailable
	}

	state, browser, verifier := token(), token(), token()
	attempt := Attempt{
		StateHash: hash(state), BrowserHash: hash(browser), Provider: provider,
		Verifier: verifier, ReturnTo: returnPath(next), ExpiresAt: time.Now().Add(LoginLifetime),
	}
	if err := s.attempts.SaveAttempt(ctx, attempt); err != nil {
		return "", "", err
	}
	return oauth.AuthorizationURL(state, verifier), browser, nil
}

func (s *LoginService) Complete(ctx context.Context, provider, browser, state, code string) (key, next string, consumed bool, err error) {
	oauth := s.providers[provider]
	if oauth == nil || browser == "" || state == "" {
		return "", "", false, ErrLoginInvalid
	}

	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	attempt, err := s.attempts.ConsumeAttempt(ctx, hash(state), hash(browser), provider)
	if err != nil {
		return "", "", false, err
	}
	if code == "" {
		return "", "", true, ErrLoginInvalid
	}

	identity, err := oauth.Identity(ctx, code, attempt.Verifier)
	if err != nil {
		return "", "", true, ErrLoginFailed
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

package session

import (
	"context"
	"fmt"
	"time"

	"github.com/grachevelloper/frets/backend/services/auth/internal/domain/auth"
)

type UseCase struct {
	repo   Repository
	tokens TokenGenerator
}

func New(repo Repository, tokens TokenGenerator) *UseCase {
	return &UseCase{
		repo:   repo,
		tokens: tokens,
	}
}

func (s *UseCase) Create(ctx context.Context, provider string, identity auth.Identity) (string, error) {
	key := s.tokens.Generate()
	if err := s.repo.CreateSession(ctx, provider, identity, auth.HashSecret(key), time.Now().Add(auth.SessionLifetime)); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return key, nil
}

func (s *UseCase) Profile(ctx context.Context, key string) (auth.Profile, error) {
	if key == "" {
		return auth.Profile{}, auth.ErrUnauthorized
	}
	profile, err := s.repo.Profile(ctx, auth.HashSecret(key))
	if err != nil {
		return auth.Profile{}, fmt.Errorf("get session profile: %w", err)
	}
	return profile, nil
}

func (s *UseCase) Revoke(ctx context.Context, key string) error {
	if key == "" {
		return auth.ErrUnauthorized
	}
	if err := s.repo.Revoke(ctx, auth.HashSecret(key)); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

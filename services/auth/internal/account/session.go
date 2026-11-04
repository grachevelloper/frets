package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"
)

const SessionLifetime = 30 * 24 * time.Hour

var ErrUnauthorized = errors.New("unauthorized")

type Profile struct {
	ID        string    `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type Identity struct {
	ID, Name string
	Email    *string
}

type Sessions interface {
	CreateSession(context.Context, string, Identity, []byte, time.Time) error
	Profile(context.Context, []byte) (Profile, error)
	Revoke(context.Context, []byte) error
}

type SessionService struct {
	repo Sessions
}

func NewSessions(repo Sessions) *SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) Create(ctx context.Context, provider string, identity Identity) (string, error) {
	key := token()
	if err := s.repo.CreateSession(ctx, provider, identity, hash(key), time.Now().Add(SessionLifetime)); err != nil {
		return "", err
	}
	return key, nil
}

func (s *SessionService) Profile(ctx context.Context, key string) (Profile, error) {
	if key == "" {
		return Profile{}, ErrUnauthorized
	}
	return s.repo.Profile(ctx, hash(key))
}

func (s *SessionService) Revoke(ctx context.Context, key string) error {
	if key == "" {
		return ErrUnauthorized
	}
	return s.repo.Revoke(ctx, hash(key))
}

func token() string {
	var data [32]byte
	rand.Read(data[:])
	return base64.RawURLEncoding.EncodeToString(data[:])
}

func hash(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

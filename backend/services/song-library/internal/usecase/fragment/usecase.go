package fragment

import (
	"context"
	"fmt"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type UseCase struct {
	repo Repository
}

func New(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) Create(ctx context.Context, owner string, fragment song.Fragment) (bool, error) {
	if err := fragment.FragmentInput.Validate(); err != nil {
		return false, err
	}
	created, err := u.repo.CreateFragment(ctx, owner, fragment)
	if err != nil {
		return false, fmt.Errorf("create fragment: %w", err)
	}
	return created, nil
}

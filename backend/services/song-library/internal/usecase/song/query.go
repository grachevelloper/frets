package song

import (
	"context"
	"fmt"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type UseCase struct {
	repo  Repository
	files Storage
}

func New(repo Repository, files Storage) *UseCase {
	return &UseCase{
		repo:  repo,
		files: files,
	}
}

func (s *UseCase) List(ctx context.Context, owner string) ([]song.Song, error) {
	items, err := s.repo.List(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("list songs: %w", err)
	}
	return items, nil
}

func (s *UseCase) GetWithFragments(ctx context.Context, owner, id string) (song.WithFragments, error) {
	item, err := s.repo.GetWithFragments(ctx, owner, id)
	if err != nil {
		return song.WithFragments{}, fmt.Errorf("get song with fragments: %w", err)
	}
	return item, nil
}

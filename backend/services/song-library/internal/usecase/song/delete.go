package song

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func (s *UseCase) Delete(ctx context.Context, owner, id string) (bool, error) {
	key, err := s.repo.BeginDeletion(ctx, owner, id)
	if err != nil {
		return false, fmt.Errorf("begin song deletion: %w", err)
	}
	if key == "" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.finishDeletion(ctx, song.Deletion{
		SongID:  id,
		FileKey: key,
	}); err != nil {
		slog.Warn("deletion deferred", "songId", id, "error", err)
		return true, nil
	}
	return false, nil
}

func (s *UseCase) finishDeletion(ctx context.Context, item song.Deletion) error {
	if err := s.files.Remove(ctx, item.FileKey); err != nil {
		return fmt.Errorf("remove song source: %w", err)
	}
	if err := s.repo.FinishDeletion(ctx, item.SongID); err != nil {
		return fmt.Errorf("finish song deletion: %w", err)
	}
	return nil
}

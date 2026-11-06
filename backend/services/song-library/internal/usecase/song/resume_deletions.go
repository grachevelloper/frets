package song

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *UseCase) ResumePendingDeletions(ctx context.Context) error {
	listCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	items, err := s.repo.PendingDeletions(listCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("list pending deletions: %w", err)
	}

	var result error
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return errors.Join(result, err)
		}
		itemCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := s.finishDeletion(itemCtx, item); err != nil {
			result = errors.Join(result, fmt.Errorf("%s: %w", item.SongID, err))
		}
		cancel()
	}
	return result
}

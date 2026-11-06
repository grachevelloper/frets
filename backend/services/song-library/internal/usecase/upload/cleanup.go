package upload

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *UseCase) CleanupExpiredAttempts(ctx context.Context) error {
	keys, err := s.repo.ExpiredKeys(ctx)
	if err != nil {
		return fmt.Errorf("list expired uploads: %w", err)
	}

	var result error
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return errors.Join(result, err)
		}
		itemCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := s.files.Remove(itemCtx, key); err != nil {
			result = errors.Join(result, fmt.Errorf("remove expired upload %s: %w", key, err))
		}
		cancel()
	}
	return result
}

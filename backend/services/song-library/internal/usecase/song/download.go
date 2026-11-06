package song

import (
	"context"
	"fmt"
	"io"
)

func (s *UseCase) Download(ctx context.Context, owner, id string) (io.ReadCloser, int64, error) {
	key, size, err := s.repo.File(ctx, owner, id)
	if err != nil {
		return nil, 0, fmt.Errorf("find song source: %w", err)
	}

	file, err := s.files.Open(ctx, key)
	if err != nil {
		return nil, 0, fmt.Errorf("open song source: %w", err)
	}
	return file, size, nil
}

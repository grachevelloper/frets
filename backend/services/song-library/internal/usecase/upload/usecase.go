package upload

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (s *UseCase) Upload(ctx context.Context, owner string, in Input) (Result, error) {
	metadata, err := song.ValidateUpload(song.Upload{
		Title: in.Title,
		Key:   in.Key,
		Hash:  in.Hash,
		Data:  in.Data,
	})
	if err != nil {
		return Result{}, err
	}

	item, attempt, err := s.repo.Reserve(ctx, owner, metadata)
	if err != nil {
		return Result{}, fmt.Errorf("reserve upload: %w", err)
	}
	if attempt == "" {
		return Result{Song: item}, nil
	}

	writeCtx, cancelWrite := context.WithTimeout(ctx, 20*time.Second)
	err = s.files.Put(writeCtx, attempt, in.Data)
	cancelWrite()
	if err != nil {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		writeErr := fmt.Errorf("write upload source: %w", err)
		if expireErr := s.repo.ExpireAttempt(releaseCtx, attempt); expireErr != nil {
			return Result{}, errors.Join(writeErr, fmt.Errorf("expire failed upload: %w", expireErr))
		}
		return Result{}, writeErr
	}

	if err := s.repo.Publish(ctx, item.ID, attempt); err != nil {
		return Result{}, fmt.Errorf("publish upload: %w", err)
	}
	return Result{
		Song:    item,
		Created: true,
	}, nil
}

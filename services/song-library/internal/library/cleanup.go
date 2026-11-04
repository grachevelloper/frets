package library

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

func (s *Service) Delete(ctx context.Context, owner, id string) (bool, error) {
	key, err := s.repo.BeginDeletion(ctx, owner, id)
	if err != nil || key == "" {
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.finishDeletion(ctx, Deletion{SongID: id, FileKey: key}); err != nil {
		log.Printf("delete %s: %v", id, err)
		return true, nil
	}
	return false, nil
}

func (s *Service) finishDeletion(ctx context.Context, item Deletion) error {
	if err := s.files.Remove(ctx, item.FileKey); err != nil {
		return err
	}
	return s.repo.FinishDeletion(ctx, item.SongID)
}

func (s *Service) Maintain(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		sweepCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		if err := s.SweepUploads(sweepCtx); err != nil {
			log.Printf("upload cleanup: %v", err)
		}
		cancel()

		deleteCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		if err := s.ResumeDeletes(deleteCtx); err != nil {
			log.Printf("delete cleanup: %v", err)
		}
		cancel()

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) SweepUploads(ctx context.Context) error {
	keys, err := s.repo.ExpiredKeys(ctx)
	if err != nil {
		return err
	}

	var result error
	for _, key := range keys {
		itemCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := s.files.Remove(itemCtx, key); err != nil {
			result = errors.Join(result, fmt.Errorf("%s: %w", key, err))
		}
		cancel()
	}
	return result
}

func (s *Service) ResumeDeletes(ctx context.Context) error {
	items, err := s.repo.PendingDeletions(ctx)
	if err != nil {
		return err
	}

	var result error
	for _, item := range items {
		itemCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := s.finishDeletion(itemCtx, item); err != nil {
			result = errors.Join(result, fmt.Errorf("%s: %w", item.SongID, err))
		}
		cancel()
	}
	return result
}

package library

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

func validTitle(title string) bool {
	return utf8.ValidString(title) && !strings.ContainsRune(title, 0) &&
		strings.TrimSpace(title) != "" && utf8.RuneCountInString(title) <= 200
}

func uploadMetadata(in Upload) (UploadMetadata, error) {
	switch {
	case in.Key == "" || len(in.Key) > 128 || !utf8.ValidString(in.Key) || strings.ContainsRune(in.Key, 0):
		return UploadMetadata{}, ErrInvalidUpload
	case !validTitle(in.Title):
		return UploadMetadata{}, ErrInvalidTitle
	case len(in.Data) == 0:
		return UploadMetadata{}, ErrInvalidUpload
	case len(in.Data) > MaxFileSize:
		return UploadMetadata{}, ErrFileTooLarge
	}

	actual := sha256.Sum256(in.Data)
	hash := hex.EncodeToString(actual[:])
	if !strings.EqualFold(hash, in.Hash) {
		return UploadMetadata{}, ErrChecksum
	}

	return UploadMetadata{Title: in.Title, Key: in.Key, Hash: hash, Size: len(in.Data)}, nil
}

func (s *Service) Upload(ctx context.Context, owner string, in Upload) (Song, bool, error) {
	metadata, err := uploadMetadata(in)
	if err != nil {
		return Song{}, false, err
	}

	song, attempt, err := s.repo.Reserve(ctx, owner, metadata)
	if err != nil || attempt == "" {
		return song, false, err
	}

	writeCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	err = s.files.Put(writeCtx, attempt, in.Data)
	cancel()
	if err != nil {
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		return Song{}, false, errors.Join(err, s.repo.ExpireAttempt(releaseCtx, attempt))
	}

	if err := s.repo.Publish(ctx, song.ID, attempt); err != nil {
		return Song{}, false, err
	}
	return song, true, nil
}

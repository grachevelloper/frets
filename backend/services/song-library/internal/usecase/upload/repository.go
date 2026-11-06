package upload

import (
	"context"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type Repository interface {
	Reserve(ctx context.Context, ownerID string, metadata song.UploadMetadata) (item song.Song, attemptID string, err error)
	Publish(ctx context.Context, songID string, attemptID string) (err error)
	ExpireAttempt(ctx context.Context, attemptID string) (err error)
	ExpiredKeys(ctx context.Context) (fileKeys []string, err error)
}

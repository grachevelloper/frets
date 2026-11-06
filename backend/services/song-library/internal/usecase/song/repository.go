package song

import (
	"context"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type Repository interface {
	List(ctx context.Context, ownerID string) (items []song.Song, err error)
	GetWithFragments(ctx context.Context, ownerID string, songID string) (item song.WithFragments, err error)
	File(ctx context.Context, ownerID string, songID string) (fileKey string, sizeBytes int64, err error)
	BeginDeletion(ctx context.Context, ownerID string, songID string) (fileKey string, err error)
	FinishDeletion(ctx context.Context, songID string) (err error)
	PendingDeletions(ctx context.Context) (items []song.Deletion, err error)
}

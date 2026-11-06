package fragment

import (
	"context"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type Repository interface {
	CreateFragment(ctx context.Context, ownerID string, fragment song.Fragment) (created bool, err error)
}

package upload

import "github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"

type Result struct {
	Song    song.Song
	Created bool
}

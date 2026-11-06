package song

import "time"

type Song struct {
	ID        string
	Title     string
	CreatedAt time.Time
}

type WithFragments struct {
	Song
	Fragments []Fragment
}

type Deletion struct {
	SongID  string
	FileKey string
}

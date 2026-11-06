package httpapi

import (
	"time"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type songDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

type fragmentDTO struct {
	ID       string `json:"id"`
	SongID   string `json:"songId"`
	Title    string `json:"title"`
	StartBar int32  `json:"startBar"`
	EndBar   int32  `json:"endBar"`
}

type fragmentRequest struct {
	Title    string `json:"title"`
	StartBar int32  `json:"startBar"`
	EndBar   int32  `json:"endBar"`
}

func toSongResponse(s song.Song) songDTO {
	return songDTO{
		ID:        s.ID,
		Title:     s.Title,
		CreatedAt: s.CreatedAt,
	}
}

func toSongListResponse(items []song.Song) []songDTO {
	response := make([]songDTO, 0, len(items))
	for _, item := range items {
		response = append(response, toSongResponse(item))
	}
	return response
}

func toFragmentResponse(f song.Fragment) fragmentDTO {
	return fragmentDTO{
		ID:       f.ID,
		SongID:   f.SongID,
		Title:    f.Title,
		StartBar: f.StartBar,
		EndBar:   f.EndBar,
	}
}

type songWithFragmentsResponse struct {
	songDTO
	Fragments []fragmentDTO `json:"fragments"`
}

func toSongWithFragmentsResponse(detail song.WithFragments) songWithFragmentsResponse {
	response := songWithFragmentsResponse{
		songDTO:   toSongResponse(detail.Song),
		Fragments: make([]fragmentDTO, 0, len(detail.Fragments)),
	}
	for _, f := range detail.Fragments {
		response.Fragments = append(response.Fragments, toFragmentResponse(f))
	}
	return response
}

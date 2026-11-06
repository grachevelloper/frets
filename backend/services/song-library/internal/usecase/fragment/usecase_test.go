package fragment

import (
	"context"
	"errors"
	"testing"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

type fragmentRepository struct {
	fragment song.Fragment
}

func (r *fragmentRepository) CreateFragment(_ context.Context, owner string, f song.Fragment) (bool, error) {
	r.fragment = f
	return true, nil
}

func TestFragmentValidationAndPreservation(t *testing.T) {
	repo := &fragmentRepository{}
	u := New(repo)
	f := song.Fragment{FragmentInput: song.FragmentInput{
		Title:    "  Припев  ",
		StartBar: 9,
		EndBar:   12,
	}}
	if created, err := u.Create(t.Context(), "owner", f); err != nil || !created || repo.fragment.Title != f.Title {
		t.Fatalf("created=%v error=%v fragment=%+v", created, err, repo.fragment)
	}
	f.StartBar = 0
	if _, err := New(nil).Create(t.Context(), "owner", f); !errors.Is(err, song.ErrInvalidFragment) {
		t.Fatalf("invalid range reached repository: %v", err)
	}
}

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grachevelloper/frets/services/song-library/internal/library"
)

func TestUploadAndDeletionPersistence(t *testing.T) {
	dsn := os.Getenv("FRETS_TEST_LIBRARY_DATABASE_URL")
	if dsn == "" {
		t.Skip("FRETS_TEST_LIBRARY_DATABASE_URL is unset")
	}
	ctx := t.Context()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)

	repo := New(db)
	owner, other := uuid.NewString(), uuid.NewString()
	in := library.UploadMetadata{Title: " Song ", Key: "request", Hash: "sha", Size: 4}
	song, attempt, err := repo.Reserve(ctx, owner, in)
	if err != nil || attempt == "" {
		t.Fatalf("reserve: song = %+v, attempt = %q, err = %v", song, attempt, err)
	}
	t.Cleanup(func() {
		for _, query := range []string{
			"DELETE FROM song_library.fragments WHERE song_id=$1",
			"DELETE FROM song_library.upload_attempts WHERE song_id=$1",
			"DELETE FROM song_library.songs WHERE id=$1",
		} {
			if _, err := db.Exec(context.Background(), query, song.ID); err != nil {
				t.Error(err)
			}
		}
	})

	if _, _, err := repo.Reserve(ctx, owner, in); !errors.Is(err, library.ErrUploadInProgress) {
		t.Fatalf("concurrent reservation: %v", err)
	}
	if _, err := db.Exec(ctx, "UPDATE song_library.upload_attempts SET expires_at=now() WHERE id=$1", attempt); err != nil {
		t.Fatal(err)
	}
	retry, newAttempt, err := repo.Reserve(ctx, owner, in)
	if err != nil || retry.ID != song.ID || newAttempt == "" || newAttempt == attempt {
		t.Fatalf("retry: song = %+v, attempt = %q, err = %v", retry, newAttempt, err)
	}
	if err := repo.Publish(ctx, song.ID, attempt); !errors.Is(err, library.ErrUploadExpired) {
		t.Fatalf("stale publication: %v", err)
	}
	if err := repo.Publish(ctx, song.ID, newAttempt); err != nil {
		t.Fatal(err)
	}
	replay, replayAttempt, err := repo.Reserve(ctx, owner, in)
	if err != nil || replay.ID != song.ID || replayAttempt != "" {
		t.Fatalf("ready replay: song = %+v, attempt = %q, err = %v", replay, replayAttempt, err)
	}

	f := library.Fragment{
		ID: uuid.NewString(), SongID: song.ID,
		FragmentInput: library.FragmentInput{Title: "Припев", StartBar: 9, EndBar: 12},
	}
	if _, err := repo.PutFragment(ctx, other, f); !errors.Is(err, library.ErrNotFound) {
		t.Fatalf("foreign fragment write: %v", err)
	}
	if created, err := repo.PutFragment(ctx, owner, f); err != nil || !created {
		t.Fatalf("fragment creation: created = %v, err = %v", created, err)
	}
	if created, err := repo.PutFragment(ctx, owner, f); err != nil || created {
		t.Fatalf("fragment replay: created = %v, err = %v", created, err)
	}
	f.EndBar++
	if _, err := repo.PutFragment(ctx, owner, f); !errors.Is(err, library.ErrFragmentConflict) {
		t.Fatalf("fragment conflict: %v", err)
	}

	if key, err := repo.BeginDeletion(ctx, owner, song.ID); err != nil || key != newAttempt {
		t.Fatalf("begin delete: key = %q, err = %v", key, err)
	}
	if _, err := repo.Detail(ctx, owner, song.ID); !errors.Is(err, library.ErrNotFound) {
		t.Fatalf("pending deletion must hide song: %v", err)
	}
	if err := repo.FinishDeletion(ctx, song.ID); err != nil {
		t.Fatal(err)
	}
	if key, err := repo.BeginDeletion(ctx, owner, song.ID); err != nil || key != "" {
		t.Fatalf("repeat delete: key = %q, err = %v", key, err)
	}
	var fragments int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM song_library.fragments WHERE song_id=$1", song.ID).Scan(&fragments); err != nil || fragments != 0 {
		t.Fatalf("fragments after deletion = %d, err = %v", fragments, err)
	}
}

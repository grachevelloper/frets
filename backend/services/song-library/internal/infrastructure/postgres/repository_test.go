package postgres

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	domainsong "github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
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
	in := domainsong.UploadMetadata{
		Title: " Song ",
		Key:   "request",
		Hash:  "sha",
		Size:  4,
	}
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

	if _, _, err := repo.Reserve(ctx, owner, in); !errors.Is(err, domainsong.ErrUploadInProgress) {
		t.Fatalf("active reservation: %v", err)
	}
	if _, err := db.Exec(ctx, "UPDATE song_library.upload_attempts SET expires_at=now() WHERE id=$1", attempt); err != nil {
		t.Fatal(err)
	}
	retry, newAttempt, err := repo.Reserve(ctx, owner, in)
	if err != nil || retry.ID != song.ID || newAttempt == "" || newAttempt == attempt {
		t.Fatalf("retry: song = %+v, attempt = %q, err = %v", retry, newAttempt, err)
	}
	if err := repo.Publish(ctx, song.ID, attempt); !errors.Is(err, domainsong.ErrUploadExpired) {
		t.Fatalf("stale publication: %v", err)
	}
	if err := repo.Publish(ctx, song.ID, newAttempt); err != nil {
		t.Fatal(err)
	}
	replay, replayAttempt, err := repo.Reserve(ctx, owner, in)
	if err != nil || replay.ID != song.ID || replayAttempt != "" {
		t.Fatalf("ready replay: song = %+v, attempt = %q, err = %v", replay, replayAttempt, err)
	}

	if items, err := repo.List(ctx, other); err != nil || len(items) != 0 {
		t.Fatalf("foreign song list: items=%v err=%v", items, err)
	}
	if _, err := repo.GetWithFragments(ctx, other, song.ID); !errors.Is(err, domainsong.ErrNotFound) {
		t.Fatalf("foreign song read: %v", err)
	}
	if _, _, err := repo.File(ctx, other, song.ID); !errors.Is(err, domainsong.ErrNotFound) {
		t.Fatalf("foreign source read: %v", err)
	}
	if _, err := repo.BeginDeletion(ctx, other, song.ID); !errors.Is(err, domainsong.ErrNotFound) {
		t.Fatalf("foreign song deletion: %v", err)
	}

	f := domainsong.Fragment{
		ID: uuid.NewString(), SongID: song.ID,
		FragmentInput: domainsong.FragmentInput{
			Title:    "Припев",
			StartBar: 9,
			EndBar:   12,
		},
	}
	if _, err := repo.CreateFragment(ctx, other, f); !errors.Is(err, domainsong.ErrNotFound) {
		t.Fatalf("foreign fragment write: %v", err)
	}
	if created, err := repo.CreateFragment(ctx, owner, f); err != nil || !created {
		t.Fatalf("fragment creation: created = %v, err = %v", created, err)
	}
	if created, err := repo.CreateFragment(ctx, owner, f); err != nil || created {
		t.Fatalf("fragment replay: created = %v, err = %v", created, err)
	}
	f.EndBar++
	if _, err := repo.CreateFragment(ctx, owner, f); !errors.Is(err, domainsong.ErrFragmentConflict) {
		t.Fatalf("fragment conflict: %v", err)
	}

	if key, err := repo.BeginDeletion(ctx, owner, song.ID); err != nil || key != newAttempt {
		t.Fatalf("begin delete: key = %q, err = %v", key, err)
	}
	if _, err := repo.GetWithFragments(ctx, owner, song.ID); !errors.Is(err, domainsong.ErrNotFound) {
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

func TestConcurrentUploadAndFragmentPersistence(t *testing.T) {
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
	owner := uuid.NewString()
	t.Cleanup(func() {
		queries := []string{
			"DELETE FROM song_library.fragments WHERE song_id IN (SELECT id FROM song_library.songs WHERE owner_id=$1)",
			"DELETE FROM song_library.upload_attempts WHERE song_id IN (SELECT id FROM song_library.songs WHERE owner_id=$1)",
			"DELETE FROM song_library.songs WHERE owner_id=$1",
		}
		for _, query := range queries {
			if _, err := db.Exec(context.Background(), query, owner); err != nil {
				t.Error(err)
			}
		}
	})
	metadata := domainsong.UploadMetadata{Title: "Concurrent song", Key: "same-key", Hash: "hash", Size: 4}
	const requests = 8
	type reservation struct {
		song    domainsong.Song
		attempt string
		err     error
	}
	results := make([]reservation, requests)
	var group sync.WaitGroup
	start := make(chan struct{})
	for i := range results {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			results[i].song, results[i].attempt, results[i].err = repo.Reserve(ctx, owner, metadata)
		}()
	}
	close(start)
	group.Wait()
	var selected reservation
	successes := 0
	for _, result := range results {
		if result.err == nil {
			successes++
			selected = result
		} else if !errors.Is(result.err, domainsong.ErrUploadInProgress) {
			t.Fatalf("unexpected reservation error: %v", result.err)
		}
	}
	if successes != 1 || selected.attempt == "" {
		t.Fatalf("successful reservations=%d selected=%+v", successes, selected)
	}
	var songs, attempts int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM song_library.songs WHERE owner_id=$1", owner).Scan(&songs); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, "SELECT count(*) FROM song_library.upload_attempts WHERE song_id=$1 AND status='active'", selected.song.ID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if songs != 1 || attempts != 1 {
		t.Fatalf("songs=%d active attempts=%d", songs, attempts)
	}
	if err := repo.Publish(ctx, selected.song.ID, selected.attempt); err != nil {
		t.Fatal(err)
	}
	fragment := domainsong.Fragment{
		ID:            uuid.NewString(),
		SongID:        selected.song.ID,
		FragmentInput: domainsong.FragmentInput{Title: "Chorus", StartBar: 9, EndBar: 12},
	}
	created := make([]bool, requests)
	failures := make([]error, requests)
	start = make(chan struct{})
	for i := range created {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			created[i], failures[i] = repo.CreateFragment(ctx, owner, fragment)
		}()
	}
	close(start)
	group.Wait()
	successes = 0
	for i, result := range created {
		if failures[i] != nil {
			t.Fatal(failures[i])
		}
		if result {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("created fragments=%d", successes)
	}
	detail, err := repo.GetWithFragments(ctx, owner, selected.song.ID)
	if err != nil || len(detail.Fragments) != 1 {
		t.Fatalf("detail=%+v err=%v", detail, err)
	}
	for i := range failures {
		group.Add(1)
		go func() {
			defer group.Done()
			f := fragment
			f.ID = uuid.NewString()
			_, failures[i] = repo.CreateFragment(ctx, owner, f)
		}()
	}
	if _, err := repo.BeginDeletion(ctx, owner, selected.song.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.FinishDeletion(ctx, selected.song.ID); err != nil {
		t.Fatal(err)
	}
	group.Wait()
	for _, err := range failures {
		if err != nil && !errors.Is(err, domainsong.ErrNotFound) {
			t.Fatal(err)
		}
	}
	var fragments int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM song_library.fragments WHERE song_id=$1", selected.song.ID).Scan(&fragments); err != nil || fragments != 0 {
		t.Fatalf("fragments after concurrent deletion=%d err=%v", fragments, err)
	}
}

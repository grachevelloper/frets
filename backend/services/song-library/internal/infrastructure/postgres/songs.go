package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/domain/song"
)

func (r *Repository) List(ctx context.Context, owner string) ([]song.Song, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id,title,created_at FROM song_library.songs
		WHERE owner_id=$1 AND status='ready' ORDER BY created_at DESC,id
	`, owner)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[song.Song])
}

func (r *Repository) GetWithFragments(ctx context.Context, owner, id string) (song.WithFragments, error) {
	var detail song.WithFragments
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return detail, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		SELECT id,title,created_at FROM song_library.songs
		WHERE id=$1 AND owner_id=$2 AND status='ready'
	`, id, owner).Scan(&detail.ID, &detail.Title, &detail.CreatedAt)
	if err != nil {
		return detail, mapRowError(err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id,title,start_bar,end_bar FROM song_library.fragments
		WHERE song_id=$1 ORDER BY start_bar,end_bar,id
	`, id)
	if err != nil {
		return detail, err
	}

	detail.Fragments, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (song.Fragment, error) {
		f := song.Fragment{SongID: id}
		err := row.Scan(&f.ID, &f.Title, &f.StartBar, &f.EndBar)
		return f, err
	})
	if err != nil {
		return detail, err
	}
	return detail, tx.Commit(ctx)
}

func (r *Repository) File(ctx context.Context, owner, id string) (string, int64, error) {
	var key string
	var size int64
	err := r.db.QueryRow(ctx, `
		SELECT file_key,size_bytes FROM song_library.songs
		WHERE id=$1 AND owner_id=$2 AND status='ready'
	`, id, owner).Scan(&key, &size)
	return key, size, mapRowError(err)
}

func (r *Repository) CreateFragment(ctx context.Context, owner string, f song.Fragment) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	status, _, err := lockSong(ctx, tx, owner, f.SongID)
	if err != nil {
		return false, err
	}
	if !status.Accessible() {
		return false, song.ErrNotFound
	}

	result, err := tx.Exec(ctx, `
		INSERT INTO song_library.fragments(id,song_id,title,start_bar,end_bar)
		VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO NOTHING
	`, f.ID, f.SongID, f.Title, f.StartBar, f.EndBar)
	if err != nil {
		return false, err
	}
	if result.RowsAffected() != 0 {
		return true, tx.Commit(ctx)
	}

	var previous song.FragmentInput
	err = tx.QueryRow(ctx, `
		SELECT title,start_bar,end_bar FROM song_library.fragments WHERE id=$1 AND song_id=$2
	`, f.ID, f.SongID).Scan(&previous.Title, &previous.StartBar, &previous.EndBar)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && f.FragmentInput != previous) {
		return false, song.ErrFragmentConflict
	}
	if err != nil {
		return false, err
	}
	return false, tx.Commit(ctx)
}

func lockSong(ctx context.Context, tx pgx.Tx, owner, id string) (song.Status, *string, error) {
	var status song.Status
	var key *string
	err := tx.QueryRow(ctx, "SELECT status,file_key FROM song_library.songs WHERE id=$1 AND owner_id=$2 FOR UPDATE", id, owner).Scan(&status, &key)
	return status, key, mapRowError(err)
}

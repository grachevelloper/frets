package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/grachevelloper/frets/services/song-library/internal/library"
)

func (r *Repository) List(ctx context.Context, owner string) ([]library.Song, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id,title,created_at FROM song_library.songs
		WHERE owner_id=$1 AND status='ready' ORDER BY created_at DESC,id
	`, owner)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[library.Song])
}

func (r *Repository) Detail(ctx context.Context, owner, id string) (library.Detail, error) {
	var detail library.Detail
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return detail, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		SELECT id,title,created_at FROM song_library.songs
		WHERE id=$1 AND owner_id=$2 AND status='ready'
	`, id, owner).Scan(&detail.ID, &detail.Title, &detail.CreatedAt)
	if err != nil {
		return detail, notFound(err)
	}

	rows, err := tx.Query(ctx, `
		SELECT id,title,start_bar,end_bar FROM song_library.fragments
		WHERE song_id=$1 ORDER BY start_bar,end_bar,id
	`, id)
	if err != nil {
		return detail, err
	}

	detail.Fragments, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (library.Fragment, error) {
		f := library.Fragment{SongID: id}
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
	return key, size, notFound(err)
}

func (r *Repository) PutFragment(ctx context.Context, owner string, f library.Fragment) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	status, _, err := lockSong(ctx, tx, owner, f.SongID)
	if err != nil {
		return false, err
	}
	if status != "ready" {
		return false, library.ErrNotFound
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

	var previous library.FragmentInput
	err = tx.QueryRow(ctx, `
		SELECT title,start_bar,end_bar FROM song_library.fragments WHERE id=$1 AND song_id=$2
	`, f.ID, f.SongID).Scan(&previous.Title, &previous.StartBar, &previous.EndBar)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && f.FragmentInput != previous) {
		return false, library.ErrFragmentConflict
	}
	if err != nil {
		return false, err
	}
	return false, tx.Commit(ctx)
}

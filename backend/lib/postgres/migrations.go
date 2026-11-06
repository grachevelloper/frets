package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, db *pgxpool.Pool, schema string, files fs.FS) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migrations: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", schema); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	table := pgx.Identifier{schema, "schema_migrations"}.Sanitize()
	if _, err := tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS "+table+" (name text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
		return fmt.Errorf("create migration history: %w", err)
	}
	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return fmt.Errorf("list migration files: %w", err)
	}
	sort.Strings(names)
	rows, err := tx.Query(ctx, "SELECT name FROM "+table+" ORDER BY name")
	if err != nil {
		return fmt.Errorf("read migration history: %w", err)
	}
	applied, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("read migration history: %w", err)
	}
	for _, name := range applied {
		if !slices.Contains(names, name) {
			return fmt.Errorf("applied migration %s is missing", name)
		}
	}
	for _, name := range names {
		if err := applyMigration(ctx, tx, table, files, name); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return tx.Commit(ctx)
}

func applyMigration(ctx context.Context, tx pgx.Tx, table string, files fs.FS, name string) error {
	data, err := fs.ReadFile(files, name)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	var previous string
	err = tx.QueryRow(ctx, "SELECT checksum FROM "+table+" WHERE name=$1", name).Scan(&previous)
	if err == nil {
		if previous != checksum {
			return fmt.Errorf("applied migration has changed")
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err := tx.Exec(ctx, string(data)); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO "+table+"(name,checksum) VALUES($1,$2)", name, checksum)
	return err
}

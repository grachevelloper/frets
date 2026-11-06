package postgres

import (
	"context"
	"os"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestMigrationsAreVersionedAndAtomic(t *testing.T) {
	dsn := os.Getenv("FRETS_TEST_POSTGRES_DATABASE_URL")
	if dsn == "" {
		t.Skip("FRETS_TEST_POSTGRES_DATABASE_URL is unset")
	}
	db, err := Open(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := "migration_test_" + uuid.NewString()[:8]
	if _, err := db.Exec(t.Context(), "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	table := pgx.Identifier{schema, "example"}.Sanitize()
	initial := "CREATE TABLE " + table + " (id integer PRIMARY KEY);"
	files := fstest.MapFS{"001_initial.sql": &fstest.MapFile{Data: []byte(initial)}}
	for range 2 {
		if err := Migrate(t.Context(), db, schema, files); err != nil {
			t.Fatal(err)
		}
	}
	delete(files, "001_initial.sql")
	if err := Migrate(t.Context(), db, schema, files); err == nil {
		t.Fatal("accepted a missing applied migration")
	}
	files["001_initial.sql"] = &fstest.MapFile{Data: []byte(initial + " -- changed")}
	if err := Migrate(t.Context(), db, schema, files); err == nil {
		t.Fatal("accepted a modified applied migration")
	}
	files["001_initial.sql"] = &fstest.MapFile{Data: []byte(initial)}
	files["002_failure.sql"] = &fstest.MapFile{Data: []byte("ALTER TABLE " + table + " ADD COLUMN value text; SELECT missing_column FROM " + table)}
	if err := Migrate(t.Context(), db, schema, files); err == nil {
		t.Fatal("accepted failed migration")
	}
	var columns int
	if err := db.QueryRow(t.Context(), "SELECT count(*) FROM information_schema.columns WHERE table_schema=$1 AND table_name='example' AND column_name='value'", schema).Scan(&columns); err != nil || columns != 0 {
		t.Fatalf("partial migration persisted: columns=%d error=%v", columns, err)
	}
}

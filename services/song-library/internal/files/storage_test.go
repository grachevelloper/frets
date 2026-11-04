package files

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestDiskStorage(t *testing.T) {
	files := diskStorage(t.TempDir())
	ctx := t.Context()
	const key = "attempt"
	const data = "original bytes"

	if err := files.Put(ctx, key, []byte(data)); err != nil {
		t.Fatal(err)
	}
	file, err := files.Open(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(file)
	file.Close()
	if err != nil || string(got) != data {
		t.Fatalf("data = %q, err = %v", got, err)
	}
	for range 2 {
		if err := files.Remove(ctx, key); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := files.Open(ctx, key); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted object: %v", err)
	}
}

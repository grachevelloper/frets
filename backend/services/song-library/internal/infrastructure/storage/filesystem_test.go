package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
)

func TestDiskStorage(t *testing.T) {
	files := Filesystem(t.TempDir())
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

func TestCanceledWriteLeavesNoObject(t *testing.T) {
	dir := t.TempDir()
	files, err := NewFilesystem(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := files.Put(ctx, "canceled", []byte("score")); !errors.Is(err, context.Canceled) {
		t.Fatalf("write error=%v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("unfinished object persisted: entries=%v error=%v", entries, err)
	}
}

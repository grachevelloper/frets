package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

func NewFilesystem(dir string) (Filesystem, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return Filesystem(dir), nil
}

type Filesystem string

func (d Filesystem) Put(ctx context.Context, key string, data []byte) error {
	file, err := os.OpenFile(filepath.Join(string(d), key+".part"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	_, writeErr := file.Write(data)
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.Rename(file.Name(), filepath.Join(string(d), key)); err != nil {
		return err
	}
	return d.syncDirectory()
}

func (d Filesystem) Open(_ context.Context, key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(string(d), key))
}

func (d Filesystem) Remove(_ context.Context, key string) error {
	for _, name := range []string{key, key + ".part"} {
		if err := os.Remove(filepath.Join(string(d), name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return d.syncDirectory()
}

func (d Filesystem) syncDirectory() error {
	dir, err := os.Open(string(d))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

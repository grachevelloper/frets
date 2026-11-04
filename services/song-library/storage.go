package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/grachevelloper/frets/services/internal/platform"
)

type storage interface {
	put(context.Context, string, []byte) error
	open(context.Context, string) (io.ReadCloser, error)
	remove(context.Context, string) error
}

func newStorage() storage {
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		dir := platform.Required("FILE_DIR")
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Fatal(err)
		}
		return diskStorage(dir)
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(platform.Required("S3_ACCESS_KEY"), platform.Required("S3_SECRET_KEY"), os.Getenv("S3_SESSION_TOKEN")),
		Secure: os.Getenv("S3_INSECURE") != "true",
		Region: os.Getenv("S3_REGION"),
	})
	if err != nil {
		log.Fatal(err)
	}
	return objectStorage{client, platform.Required("S3_BUCKET")}
}

type diskStorage string

func (d diskStorage) put(ctx context.Context, key string, data []byte) error {
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

func (d diskStorage) open(_ context.Context, key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(string(d), key))
}

func (d diskStorage) remove(_ context.Context, key string) error {
	for _, name := range []string{key, key + ".part"} {
		if err := os.Remove(filepath.Join(string(d), name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return d.syncDirectory()
}

func (d diskStorage) syncDirectory() error {
	dir, err := os.Open(string(d))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

type objectStorage struct {
	client *minio.Client
	bucket string
}

func (s objectStorage) put(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream", DisableMultipart: true,
	})
	return err
}

func (s objectStorage) open(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	if _, err := object.Stat(); err != nil {
		object.Close()
		return nil, err
	}
	return object, nil
}

func (s objectStorage) remove(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil && strings.EqualFold(minio.ToErrorResponse(err).Code, "NoSuchKey") {
		return nil
	}
	return err
}

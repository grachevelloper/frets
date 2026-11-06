package storage

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Config struct {
	Endpoint     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	SessionToken string
	Region       string
	Insecure     bool
}

func NewS3(config S3Config) (*S3, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, config.SessionToken),
		Secure: !config.Insecure,
		Region: config.Region,
	})
	if err != nil {
		return nil, err
	}
	return &S3{
		client: client,
		bucket: config.Bucket,
	}, nil
}

type S3 struct {
	client *minio.Client
	bucket string
}

func (s S3) Put(ctx context.Context, key string, data []byte) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream", DisableMultipart: true,
	})
	return err
}

func (s S3) Open(ctx context.Context, key string) (io.ReadCloser, error) {
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

func (s S3) Remove(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil && strings.EqualFold(minio.ToErrorResponse(err).Code, "NoSuchKey") {
		return nil
	}
	return err
}

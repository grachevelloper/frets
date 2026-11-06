package song

import (
	"context"
	"io"
)

type Storage interface {
	Open(ctx context.Context, fileKey string) (file io.ReadCloser, err error)
	Remove(ctx context.Context, fileKey string) (err error)
}

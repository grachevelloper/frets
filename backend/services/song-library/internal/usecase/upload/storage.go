package upload

import (
	"context"
)

type Storage interface {
	Put(ctx context.Context, fileKey string, data []byte) (err error)
	Remove(ctx context.Context, fileKey string) (err error)
}

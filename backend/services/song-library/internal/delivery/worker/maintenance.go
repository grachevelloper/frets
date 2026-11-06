package worker

import (
	"context"
	"log/slog"
	"time"
)

type UploadMaintenance interface {
	CleanupExpiredAttempts(ctx context.Context) (err error)
}

type SongMaintenance interface {
	ResumePendingDeletions(ctx context.Context) (err error)
}

func Run(ctx context.Context, uploads UploadMaintenance, songs SongMaintenance) {
	uploadsStopped := make(chan struct{})
	go func() {
		defer close(uploadsStopped)
		run(ctx, "upload cleanup", func(ctx context.Context) error {
			sweepCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			return uploads.CleanupExpiredAttempts(sweepCtx)
		})
	}()

	run(ctx, "delete cleanup", songs.ResumePendingDeletions)
	<-uploadsStopped
}

func run(ctx context.Context, operation string, maintain func(context.Context) error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}

		if err := maintain(ctx); err != nil && ctx.Err() == nil {
			slog.Error(operation, "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

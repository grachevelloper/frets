package worker

import (
	"context"
	"testing"
	"time"
)

type uploadMaintenanceStub struct {
	calls chan struct{}
}

func (u uploadMaintenanceStub) CleanupExpiredAttempts(context.Context) error {
	u.calls <- struct{}{}
	return nil
}

type blockedSongMaintenance struct{}

func (blockedSongMaintenance) ResumePendingDeletions(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestSlowDeletionDoesNotBlockUploadCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls := make(chan struct{}, 2)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		Run(ctx, uploadMaintenanceStub{calls: calls}, blockedSongMaintenance{})
	}()
	deadline := time.NewTimer(7 * time.Second)
	defer deadline.Stop()
	for range 2 {
		select {
		case <-calls:
		case <-deadline.C:
			t.Fatal("slow deletion blocked recurring upload cleanup")
		}
	}
	cancel()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("maintenance did not stop after cancellation")
	}
}

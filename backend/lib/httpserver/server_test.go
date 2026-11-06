package httpserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRunStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan error, 1)
	go func() { stopped <- Run(ctx, "127.0.0.1:0", http.NewServeMux()) }()
	cancel()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP server did not stop")
	}
}

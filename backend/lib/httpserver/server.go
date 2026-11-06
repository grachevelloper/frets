package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

func Run(ctx context.Context, address string, handler http.Handler) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       40 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	stopped := make(chan error, 1)
	go func() { stopped <- server.Serve(listener) }()
	select {
	case err := <-stopped:
		return err
	case <-ctx.Done():
		// Shutdown needs a fresh context because the application context is already canceled.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return err
		}
		if err := <-stopped; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

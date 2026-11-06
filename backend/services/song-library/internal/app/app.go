package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grachevelloper/frets/backend/lib/httpserver"
	db "github.com/grachevelloper/frets/backend/lib/postgres"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/config"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/delivery"
	httpapi "github.com/grachevelloper/frets/backend/services/song-library/internal/delivery/http"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/delivery/worker"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/infrastructure/authclient"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/infrastructure/postgres"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/infrastructure/storage"
	fragments "github.com/grachevelloper/frets/backend/services/song-library/internal/usecase/fragment"
	songs "github.com/grachevelloper/frets/backend/services/song-library/internal/usecase/song"
	"github.com/grachevelloper/frets/backend/services/song-library/internal/usecase/upload"
	"github.com/grachevelloper/frets/backend/services/song-library/migrations"
)

type App struct {
	db      *pgxpool.Pool
	address string
	handler http.Handler
	songs   *songs.UseCase
	uploads *upload.UseCase
}

type fileStorage interface {
	songs.Storage
	upload.Storage
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	pool, err := db.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(startupCtx, pool, "song_library", migrations.Files); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate song library: %w", err)
	}
	files, err := openStorage(cfg)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("open storage: %w", err)
	}
	repo := postgres.New(pool)
	songQueries := songs.New(repo, files)
	uploads := upload.New(repo, files)
	bookmarks := fragments.New(repo)
	sessions := authclient.New(cfg.AuthURL, cfg.InternalSecret)
	handler := httpapi.New(songQueries, uploads, bookmarks)
	return &App{
		db:      pool,
		address: cfg.Address,
		handler: delivery.NewRouter(cfg.PublicOrigin, handler, sessions),
		songs:   songQueries,
		uploads: uploads,
	}, nil
}

func openStorage(cfg config.Config) (fileStorage, error) {
	if cfg.S3.Endpoint == "" {
		return storage.NewFilesystem(cfg.FileDir)
	}
	return storage.NewS3(storage.S3Config{
		Endpoint:     cfg.S3.Endpoint,
		Bucket:       cfg.S3.Bucket,
		AccessKey:    cfg.S3.AccessKey,
		SecretKey:    cfg.S3.SecretKey,
		SessionToken: cfg.S3.SessionToken,
		Region:       cfg.S3.Region,
		Insecure:     cfg.S3.Insecure,
	})
}

func (a *App) Run(ctx context.Context) error {
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		worker.Run(workerCtx, a.uploads, a.songs)
	}()
	err := httpserver.Run(ctx, a.address, a.handler)
	cancel()
	<-stopped
	return err
}

func (a *App) Close() {
	a.db.Close()
}

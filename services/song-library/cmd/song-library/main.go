package main

import (
	"context"
	"net/http"

	"github.com/grachevelloper/frets/services/internal/platform"
	"github.com/grachevelloper/frets/services/song-library/internal/authclient"
	"github.com/grachevelloper/frets/services/song-library/internal/files"
	"github.com/grachevelloper/frets/services/song-library/internal/httpapi"
	"github.com/grachevelloper/frets/services/song-library/internal/library"
	"github.com/grachevelloper/frets/services/song-library/internal/postgres"
)

func main() {
	db := platform.Database("song_library", postgres.Schema)
	service := library.New(postgres.New(db), files.FromEnvironment())
	sessions := authclient.New(platform.Required("AUTH_URL"), platform.Required("INTERNAL_SECRET"))
	go service.Maintain(context.Background())

	mux := http.NewServeMux()
	httpapi.Register(mux, service, sessions)
	platform.Serve(mux)
}

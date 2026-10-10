package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PhilippSoeder/song-guessing-game/internal/config"
	"github.com/PhilippSoeder/song-guessing-game/internal/server"
	"github.com/PhilippSoeder/song-guessing-game/internal/spotify"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	spotifyClient := spotify.NewClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRedirectURI)
	router := server.NewRouter(cfg, spotifyClient)

	fmt.Printf("Server running on http://localhost:%s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

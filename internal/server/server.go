package server

import (
	"net/http"
	"strings"

	"github.com/PhilippSoeder/song-guessing-game/internal/config"
	"github.com/PhilippSoeder/song-guessing-game/internal/spotify"
)

type server struct {
	cfg     *config.Config
	spotify *spotify.Client
}

func NewRouter(cfg *config.Config, spotifyClient *spotify.Client) http.Handler {
	router := http.NewServeMux()
	s := &server{cfg: cfg, spotify: spotifyClient}

	router.HandleFunc("/", s.serveIndex)
	router.Handle("/static/", noStore(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
	router.HandleFunc("/api/auth", s.handleSpotifyLogin)
	router.HandleFunc("/api/auth-url", s.handleAuthURL)

	registeredCallbackPaths := make(map[string]struct{})
	registerCallback := func(path string) {
		path = strings.TrimRight(path, "/")
		if path == "" {
			path = "/"
		}

		for _, callbackPath := range []string{path, path + "/"} {
			if _, exists := registeredCallbackPaths[callbackPath]; exists {
				continue
			}
			router.HandleFunc(callbackPath, s.handleSpotifyCallback)
			registeredCallbackPaths[callbackPath] = struct{}{}
		}
	}

	// Register configured callback path and compatibility aliases.
	registerCallback(cfg.SpotifyCallbackPath)
	registerCallback("/api/spotify-callback")
	registerCallback("/spotify-callback")
	router.HandleFunc("/api/session", s.handleSession)
	router.HandleFunc("/api/spotify-token", s.handleSpotifyToken)
	router.HandleFunc("/api/token-refresh", s.handleTokenRefresh)
	router.HandleFunc("/api/track-info", s.handleTrackInfo)

	return router
}

func (s *server) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

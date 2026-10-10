package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PhilippSoeder/song-guessing-game/internal/config"
	"github.com/PhilippSoeder/song-guessing-game/internal/spotify"
)

func TestSpotifyLoginRedirectsToSpotify(t *testing.T) {
	router := NewRouter(
		&config.Config{SpotifyCallbackPath: "/api/spotify-callback"},
		spotify.NewClient("client-id", "client-secret", "https://example.ts.net/api/spotify-callback"),
	)
	request := httptest.NewRequest(http.MethodGet, "/api/auth", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("login status = %d, want %d", response.Code, http.StatusSeeOther)
	}
	if location := response.Header().Get("Location"); !strings.HasPrefix(location, "https://accounts.spotify.com/authorize?") {
		t.Fatalf("login redirect location = %q, want Spotify authorize URL", location)
	}
}

func TestSessionResponseIsNotCached(t *testing.T) {
	router := NewRouter(
		&config.Config{SpotifyCallbackPath: "/api/spotify-callback"},
		spotify.NewClient("client-id", "client-secret", "https://example.ts.net/api/spotify-callback"),
	)
	request := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want %q", got, "no-store")
	}
}

func TestPageAndStaticAssetsAreNotCached(t *testing.T) {
	t.Chdir("../..")
	router := NewRouter(
		&config.Config{SpotifyCallbackPath: "/api/spotify-callback"},
		spotify.NewClient("client-id", "client-secret", "https://example.ts.net/api/spotify-callback"),
	)
	for _, path := range []string{"/", "/static/js/app.js"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if got := response.Result().Header.Get("Cache-Control"); got != "no-store" {
				t.Fatalf("Cache-Control = %q, want %q (status %d)", got, "no-store", response.Code)
			}
			if path == "/" && (!strings.Contains(response.Body.String(), "styles.css?v=20261010") || !strings.Contains(response.Body.String(), "app.js?v=20261010")) {
				t.Fatal("index page is missing versioned static asset URLs")
			}
		})
	}
}

func TestNewRouterRegistersSpotifyCallbackOnce(t *testing.T) {
	for _, callbackPath := range []string{
		"/api/spotify-callback",
		"/api/spotify-callback/",
	} {
		t.Run(callbackPath, func(t *testing.T) {
			router := NewRouter(
				&config.Config{SpotifyCallbackPath: callbackPath},
				spotify.NewClient("client-id", "client-secret", "https://example.ts.net"+callbackPath),
			)

			request := httptest.NewRequest(http.MethodGet, "/api/spotify-callback", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("callback status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestParseSpotifyTrackID(t *testing.T) {
	const trackID = "2RZWdE8kYPlCAcRUYDeuLC"
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "raw ID", input: trackID, want: trackID},
		{name: "Spotify URI", input: "spotify:track:" + trackID, want: trackID},
		{name: "share URL", input: "https://open.spotify.com/track/" + trackID, want: trackID},
		{name: "localized share URL", input: "https://open.spotify.com/intl-de/track/" + trackID, want: trackID},
		{name: "localized share URL with share query", input: "https://open.spotify.com/intl-de/track/" + trackID + "?si=50f60b1c74ff4447", want: trackID},
		{name: "unsupported host", input: "https://notspotify.com/intl-de/track/" + trackID, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseSpotifyTrackID(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseSpotifyTrackID() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("parseSpotifyTrackID() = %q, want %q", got, test.want)
			}
		})
	}
}

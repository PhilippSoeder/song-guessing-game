package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var trackIDRegex = regexp.MustCompile(`^[A-Za-z0-9]{22}$`)

func (s *server) handleAuthURL(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, AuthURLResponse{AuthURL: s.spotify.BuildAuthURL()})
}

func (s *server) handleSpotifyLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.spotify.BuildAuthURL(), http.StatusSeeOther)
}

func (s *server) handleSpotifyCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	_, err := s.spotify.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?v=20261010", http.StatusSeeOther)
}

func (s *server) handleSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	_, expiresIn, authenticated := s.spotify.CurrentAccessToken()
	if !authenticated {
		expiresIn = 0
	}
	writeJSON(w, http.StatusOK, SessionResponse{Authenticated: authenticated, ExpiresIn: expiresIn})
}

func (s *server) handleSpotifyToken(w http.ResponseWriter, r *http.Request) {
	accessToken, expiresIn, err := s.spotify.GetAccessToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{AccessToken: accessToken, ExpiresIn: expiresIn})
}

func (s *server) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	accessToken, expiresIn, err := s.spotify.RefreshAccessToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{AccessToken: accessToken, ExpiresIn: expiresIn})
}

func (s *server) handleTrackInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TrackInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	trackID, err := parseSpotifyTrackID(req.QRCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	track, err := s.spotify.FetchTrackMetadata(trackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TrackInfoResponse{
		TrackURI:    fmt.Sprintf("spotify:track:%s", trackID),
		TrackName:   track.Name,
		ArtistName:  firstArtist(track.Artists),
		AlbumName:   track.Album.Name,
		ReleaseDate: track.Album.ReleaseDate,
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func parseSpotifyTrackID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if trackIDRegex.MatchString(raw) {
		return raw, nil
	}

	if strings.HasPrefix(raw, "spotify:track:") {
		trackID := strings.TrimPrefix(raw, "spotify:track:")
		if trackIDRegex.MatchString(trackID) {
			return trackID, nil
		}
	}

	parsed, err := url.Parse(raw)
	if err == nil {
		host := strings.ToLower(parsed.Hostname())
		if host == "open.spotify.com" || host == "www.open.spotify.com" {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			for index, part := range parts {
				if part == "track" && index+1 < len(parts) && trackIDRegex.MatchString(parts[index+1]) {
					return parts[index+1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("unsupported QR code payload: expected Spotify track ID, URI, or track URL")
}

func firstArtist(artists []struct {
	Name string `json:"name"`
}) string {
	if len(artists) == 0 {
		return "Unknown Artist"
	}
	return artists[0].Name
}

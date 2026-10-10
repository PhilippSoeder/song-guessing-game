package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	SpotifyCallbackPath string
	Port                string
}

func Load() (*Config, error) {
	cfg := &Config{
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:  os.Getenv("SPOTIFY_REDIRECT_URI"),
		Port:                os.Getenv("PORT"),
	}

	if cfg.SpotifyClientID == "" || cfg.SpotifyClientSecret == "" || cfg.SpotifyRedirectURI == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET, and SPOTIFY_REDIRECT_URI must be set")
	}
	if cfg.Port == "" {
		cfg.Port = "8888"
	}

	callbackPath, err := parseCallbackPath(cfg.SpotifyRedirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid SPOTIFY_REDIRECT_URI: %w", err)
	}
	if callbackPath == "/" {
		return nil, fmt.Errorf("SPOTIFY_REDIRECT_URI must include a callback path, not root /")
	}

	cfg.SpotifyCallbackPath = callbackPath
	return cfg, nil
}

func parseCallbackPath(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if parsed.Path == "" {
		return "/", nil
	}
	return parsed.Path, nil
}

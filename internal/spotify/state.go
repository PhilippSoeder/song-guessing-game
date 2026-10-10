package spotify

import (
	"sync"
	"time"
)

type TokenStore struct {
	sync.Mutex
	accessToken   string
	refreshToken  string
	expiry        time.Time
	authenticated bool
}

func (s *TokenStore) Update(token spotifyTokenResponse) {
	s.Lock()
	defer s.Unlock()

	s.accessToken = token.AccessToken
	if token.RefreshToken != "" {
		s.refreshToken = token.RefreshToken
	}
	s.expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	s.authenticated = true
}

func (s *TokenStore) CurrentAccessToken() (string, int, bool) {
	s.Lock()
	defer s.Unlock()

	if !s.authenticated {
		return "", 0, false
	}
	expiresIn := int(time.Until(s.expiry).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}
	return s.accessToken, expiresIn, true
}

func (s *TokenStore) RefreshToken() string {
	s.Lock()
	defer s.Unlock()
	return s.refreshToken
}

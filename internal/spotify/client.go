package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	clientID     string
	clientSecret string
	redirectURI  string
	store        *TokenStore
}

func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		store:        &TokenStore{},
	}
}

func (c *Client) BuildAuthURL() string {
	scopes := []string{
		"streaming",
		"user-read-playback-state",
		"user-modify-playback-state",
		"user-read-private",
		"user-read-email",
	}
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", c.redirectURI)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("show_dialog", "true")
	return "https://accounts.spotify.com/authorize?" + params.Encode()
}

func (c *Client) ExchangeCodeForToken(code string) (*spotifyTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.redirectURI)

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth(c.clientID, c.clientSecret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spotify token exchange failed: %s", string(body))
	}

	var tokenResp spotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	c.store.Update(tokenResp)
	return &tokenResp, nil
}

func (c *Client) RefreshAccessToken() (string, int, error) {
	refreshToken := c.store.RefreshToken()
	if refreshToken == "" {
		return "", 0, fmt.Errorf("no refresh token available")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth(c.clientID, c.clientSecret))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("spotify refresh failed: %s", string(body))
	}

	var tokenResp spotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, err
	}

	c.store.Update(tokenResp)
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

func (c *Client) GetAccessToken() (string, int, error) {
	accessToken, expiresIn, authenticated := c.store.CurrentAccessToken()
	if !authenticated {
		return "", 0, fmt.Errorf("not authenticated")
	}
	if expiresIn < 10 {
		return c.RefreshAccessToken()
	}
	return accessToken, expiresIn, nil
}

func (c *Client) CurrentAccessToken() (string, int, bool) {
	return c.store.CurrentAccessToken()
}

func (c *Client) FetchTrackMetadata(trackID string) (*spotifyTrackResponse, error) {
	accessToken, _, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", url.PathEscape(trackID))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		accessToken, _, err = c.RefreshAccessToken()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spotify fetch track failed: %s", string(body))
	}

	var track spotifyTrackResponse
	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		return nil, err
	}
	return &track, nil
}

func basicAuth(clientID, secret string) string {
	credentials := clientID + ":" + secret
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}

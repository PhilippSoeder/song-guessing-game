package server

type TrackInfoRequest struct {
	QRCode string `json:"qr_code"`
}

type AuthURLResponse struct {
	AuthURL string `json:"auth_url"`
}

type SessionResponse struct {
	Authenticated bool `json:"authenticated"`
	ExpiresIn     int  `json:"expires_in"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type TrackInfoResponse struct {
	TrackURI    string `json:"track_uri"`
	TrackName   string `json:"track_name"`
	ArtistName  string `json:"artist_name"`
	AlbumName   string `json:"album_name"`
	ReleaseDate string `json:"release_date"`
}

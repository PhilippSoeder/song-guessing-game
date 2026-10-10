# Song Guessing Game

## Overview

This mobile browser game scans Spotify track QR cards, plays the selected song through the Spotify Web Playback SDK, and reveals song metadata on demand. The backend is written in Go; the frontend uses plain HTML, CSS, and JavaScript.

This app is intentionally designed for a single deployment, single host, single backend instance. It is not meant to be a multi-user, multi-tenant, or horizontally scaled web application. OAuth tokens are held in process memory and are lost when the server restarts. The app does not provide separate user sessions, account isolation, or a hardened production auth layer.

This project keeps the implementation intentionally small and pragmatic. The named downsides below are therefore not addressed yet, because the app is not designed as a shared or production-grade service. If there is time and motivation later, some or all of these concerns could be tackled in a future iteration.

This project is an independent, non-commercial, fan-made implementation inspired by the gameplay concept of Hitster by Jumbo Group.

It is not affiliated with, endorsed by, or sponsored by Jumbo Group. This project does not use Jumbo Group branding, logos, official assets, or copyrighted materials from the original game. The software is intended for personal, educational, and non-commercial use only.

## Requirements

* Go 1.27.2 or later
* Docker engine for container packaging
* Spotify Developer credentials (`SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET`)
* A Spotify account with Web Playback support (Spotify Premium is typically required for the Web Playback SDK)
* A modern mobile browser with camera permissions
* For mobile use, a reachable HTTPS origin with a certificate trusted by the phone

## Supported QR payloads

The app accepts these Spotify track formats in QR codes:

* raw Spotify track ID, e.g. `3n3Ppam7vgaVa1iaRUc9Lp`
* Spotify track URI, e.g. `spotify:track:3n3Ppam7vgaVa1iaRUc9Lp`
* Open Spotify track URL, e.g. `https://open.spotify.com/track/3n3Ppam7vgaVa1iaRUc9Lp`

## Playing a Game

1. Select **Authenticate with Spotify** and approve access.
2. After returning to the app, select **Start Scanning** and allow camera access.
3. Scan a track QR code. Playback starts automatically while song details stay hidden.
4. Use the playback controls to pause, resume, or restart the track.
5. Select **Reveal** to show the song details. Select **Hide** to conceal them again.
6. Select **Next** to scan another card and begin the next round.

## Environment variables

Set these before running the app:

* `SPOTIFY_CLIENT_ID`
* `SPOTIFY_CLIENT_SECRET`
* `SPOTIFY_REDIRECT_URI`
* `PORT` (optional, default: `8888`)

The redirect URI must exactly match the URI configured in your Spotify Developer Dashboard and point to `/api/spotify-callback` on the host running the app. Use the same reachable hostname for the page and callback. A `localhost` redirect works for local browser development, but on a phone `localhost` points to the phone itself.

## Local development

Build and run locally:

```bash
export SPOTIFY_CLIENT_ID="your-client-id"
export SPOTIFY_CLIENT_SECRET="your-client-secret"
export SPOTIFY_REDIRECT_URI="http://localhost:8888/api/spotify-callback"
export PORT=8888

go run ./cmd/song-guessing-game
```

Then open `http://localhost:8888` in a browser on the development computer. This HTTP localhost setup is for local development; it does not provide camera access to a phone on the network.

## Mobile Access

To use the app from a phone, expose the Go server through an HTTPS reverse proxy or another HTTPS-capable host reachable by the phone. Set `SPOTIFY_REDIRECT_URI` to that same HTTPS host followed by `/api/spotify-callback`, register the exact URI in the Spotify Developer Dashboard, and open the HTTPS app URL on the phone. The certificate must be trusted by the phone for camera access to work.

## Makefile Commands

The Makefile provides shortcuts for building and pushing the Docker image:

* `make build` builds the image locally and loads it into Docker.
* `make` pushes the image to the configured registry. This is equivalent to `make push`.

Set the registry when invoking Make:

```bash
make REGISTRY=ghcr.io/OWNER
```

To keep a default registry for your own checkout, create a local `Makefile.local` file:

```make
REGISTRY := ghcr.io/OWNER
```

`Makefile.local` is ignored by Git, so keep personal hostnames and registry settings there rather than in the tracked Makefile. A fresh clone must set `REGISTRY` on the command line or create its own `Makefile.local` before running `make`.

## Docker

Build the image:

```bash
docker build -t song-guessing-game-app .
```

Run the container:

```bash
docker run --rm -p 8888:8888 \
  -e SPOTIFY_CLIENT_ID="your-client-id" \
  -e SPOTIFY_CLIENT_SECRET="your-client-secret" \
  -e SPOTIFY_REDIRECT_URI="http://localhost:8888/api/spotify-callback" \
  song-guessing-game-app
```

This example is for local development on the computer running Docker. For mobile access, use the HTTPS setup above and set `SPOTIFY_REDIRECT_URI` to the publicly or locally reachable HTTPS callback URL.

## Security and design trade-offs

This project is intentionally scoped to a single deployment and a single host. That makes the implementation simple and workable, but it also means several security and robustness concerns are not addressed yet:

* OAuth tokens are stored in process memory only and are lost on restart; they are not persisted or encrypted.
* There is no user/session isolation. Anyone who can access the running app effectively shares the same backend state.
* There is no persistent session store, no per-user login state, and no multi-user access model.
* The callback flow does not implement advanced protections such as OAuth state validation or explicit CSRF defense beyond the basic app scope.
* The app assumes a trusted single browser/client environment and a single backend instance; it is not designed for public internet exposure or shared deployments.
* There is no load balancing, replica coordination, or distributed token/session management.

This is an explicit trade-off for a small single-host game app. The project is not attempting to solve these concerns yet, and they are intentionally left out because the app is meant to stay narrow and easy to operate. A future version might add better state management, stronger OAuth protections, persistence, and more robust deployment boundaries if needed.

## Notes

* OAuth tokens are held in memory and are not persisted across server restarts.
* The app is intended for one host and one backend process; it does not isolate multiple users or replicas.
* Spotify playback depends on browser support and account permissions. Spotify Premium may be required for the Web Playback SDK.
* If playback does not start, confirm the account permissions and browser support. Camera access requires permission and a secure browser origin.

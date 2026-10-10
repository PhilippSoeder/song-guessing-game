---
title: Song Guessing Game Requirements
description: User journey and functional requirements for the Spotify QR-code song guessing game
---

## User Journey

1. The host opens the game in a mobile browser using a URL reachable from the phone. Camera access requires HTTPS or a browser-recognized localhost origin.
2. The host selects "Authenticate with Spotify" and approves access with their Spotify account.
3. After returning from Spotify, the host selects "Start Scanning" and grants camera access if prompted.
4. The host scans a card containing a Spotify track ID, URI, or track URL.
5. The camera stops and playback controls appear. The scanned song starts automatically, while its metadata remains hidden.
6. The host can pause, resume, or restart the current song. There is no control for skipping to an arbitrary track.
7. The host can select "Reveal" to show the song, artist, album, and release date, then select "Hide" to conceal them again.
8. The host selects "Next" to start another scan and repeat the round without reloading the page.

## Requirements

* The game is a single-host web application. The host opens it in a mobile browser without installing an app.
* The host's phone camera scans QR codes containing Spotify track IDs, URIs, or track URLs.
* The Go backend handles Spotify OAuth, token exchange and refresh, and track metadata lookup. The browser uses the Spotify Web Playback SDK and Spotify Web API for playback.
* Mobile camera access and Spotify OAuth must use a reachable origin. For phones on the local network, use HTTPS with a certificate trusted by the phone; `localhost` refers to the phone itself, not the computer running the app.
* The OAuth redirect URI must use the same reachable host and callback path as configured in the Spotify Developer Dashboard.
* The frontend uses plain HTML, CSS, and JavaScript without a framework, and adapts to mobile screens.
* Song metadata stays hidden until the host selects "Reveal" and can be hidden again with "Hide".
* The host can play, pause, resume, and restart the current song, then scan another card to continue. Playback depends on browser support and Spotify account permissions; Spotify Premium may be required for Web Playback.
* This deployment is intended for one host and one running backend process. OAuth tokens are kept in process memory, are not persisted across restarts, and the app does not provide separate user sessions.

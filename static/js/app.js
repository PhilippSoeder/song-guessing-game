const authSection = document.getElementById('auth-section');

const gameSection = document.getElementById('game-section');
const startScanBtn = document.getElementById('start-scan-btn');
const playbackControls = document.getElementById('playback-controls');
const nextControls = document.getElementById('next-controls');
const revealControls = document.getElementById('reveal-controls');
const revealBtn = document.getElementById('reveal-btn');
const nextBtn = document.getElementById('next-btn');
const playPauseBtn = document.getElementById('play-pause-btn');
const restartBtn = document.getElementById('restart-btn');

const trackTitle = document.getElementById('track-title');
const trackArtist = document.getElementById('track-artist');
const trackAlbum = document.getElementById('track-album');
const trackRelease = document.getElementById('track-release');
const readerDiv = document.getElementById('reader');
const uiNotice = document.getElementById('ui-notice');

let player = null;
let deviceId = null;
let accessToken = null;
let currentTrack = null;
let isPlaying = false;
let scanInterval = null;
let videoStream = null;

startScanBtn.addEventListener('click', startScanFromGesture);

revealBtn.addEventListener('click', () => {
    if (!currentTrack) return;
    const info = document.getElementById('track-info');
    if (info && info.style.display === 'block') {
        info.style.display = 'none';
        revealBtn.textContent = 'Reveal';
        return;
    }

    trackTitle.textContent = currentTrack.track_name;
    trackArtist.textContent = currentTrack.artist_name;
    trackAlbum.textContent = currentTrack.album_name;
    trackRelease.textContent = currentTrack.release_date || 'Unknown';
    trackTitle.classList.remove('hidden');
    trackArtist.classList.remove('hidden');
    trackAlbum.classList.remove('hidden');
    trackRelease.classList.remove('hidden');
    if (info) info.style.display = 'block';
    revealBtn.textContent = 'Hide';
});

nextBtn.addEventListener('click', async () => {
    await activatePlayerFromGesture();
    try { await pauseTrack(); } catch (e) { /* ignore */ }
    resetRound();
    await beginScan();
});

playPauseBtn.addEventListener('click', async () => {
    if (!currentTrack || !deviceId) return;
    await activatePlayerFromGesture();
    if (isPlaying) {
        await pauseTrack();
    } else {
        await resumeTrack();
    }
});

restartBtn.addEventListener('click', async () => {
    if (!currentTrack || !deviceId) return;
    await activatePlayerFromGesture();
    await playTrack(currentTrack.track_uri, true);
});

window.addEventListener('load', init);

async function init() {
    try {
        const session = await fetchJSON('/api/session', { cache: 'no-store' });
        if (!session.authenticated) {
            return;
        }

        authSection.style.display = 'none';
        gameSection.style.display = 'block';
        const playerReady = await setupSpotifyPlayer();
        if (!playerReady) return;

        clearNotice();
        startScanBtn.disabled = false;
    } catch (error) {
        showNotice('Unable to verify Spotify session. Reload the page and try again.', true);
        console.error(error);
    }
}

async function fetchJSON(url, options = {}) {
    const res = await fetch(url, options);
    if (!res.ok) {
        const content = await res.text();
        throw new Error(content || res.statusText);
    }
    return res.json();
}

function showNotice(message, isError = false) {
    uiNotice.textContent = message;
    uiNotice.style.display = 'block';
    uiNotice.classList.toggle('error', isError);
}

function clearNotice() {
    uiNotice.style.display = 'none';
    uiNotice.textContent = '';
}

function updateControlStates({ playbackEnabled = false } = {}) {
    playPauseBtn.disabled = !playbackEnabled;
    restartBtn.disabled = !playbackEnabled;
    revealBtn.disabled = !currentTrack;
    nextBtn.disabled = !currentTrack;
}

async function setupSpotifyPlayer() {
    try {
        await loadSpotifySDK();
        player = new Spotify.Player({
            name: 'Song Guessing Game',
            getOAuthToken: async (cb) => {
                const tokenData = await fetchJSON('/api/spotify-token');
                accessToken = tokenData.access_token;
                cb(accessToken);
            },
            volume: 0.95,
        });

        let resolveReady;
        const readyPromise = new Promise((resolve) => {
            resolveReady = resolve;
        });

        player.addListener('ready', ({ device_id }) => {
            deviceId = device_id;
            updateControlStates({ playbackEnabled: false });
            clearNotice();
            resolveReady(device_id);
        });

        player.addListener('not_ready', ({ device_id }) => {
            console.warn('Spotify device went offline:', device_id);
            if (deviceId === device_id) deviceId = null;
            showNotice('Spotify playback device went offline. Reload the page or sign in again.', true);
            updateControlStates({ playbackEnabled: false });
        });

        player.addListener('initialization_error', ({ message }) => {
            console.error('Spotify initialization error:', message);
            showNotice('Spotify SDK failed to initialize. Make sure your browser supports playback SDK.', true);
        });

        player.addListener('authentication_error', ({ message }) => {
            console.error('Spotify auth error:', message);
            showNotice('Spotify authentication error occurred. Sign in again and refresh the page.', true);
        });

        player.addListener('account_error', ({ message }) => {
            console.error('Spotify account error:', message);
            showNotice('Spotify account error detected. Confirm the account has playback permissions.', true);
        });

        player.addListener('player_state_changed', (state) => {
            if (!state) {
                return;
            }
            isPlaying = !state.paused;
            playPauseBtn.textContent = state.paused ? '▶' : 'Ⅱ';
        });

        player.addListener('autoplay_failed', () => {
            isPlaying = false;
            playPauseBtn.textContent = '▶';
            showNotice('Tap a playback control in this page to allow audio, then try again.', true);
        });

        player.addListener('playback_error', ({ message }) => {
            console.error('Spotify playback error:', message);
            showNotice(`Spotify playback error: ${message}`, true);
        });

        const connected = await player.connect();
        if (!connected) {
            throw new Error('Could not connect Spotify player');
        }

        let timeoutId;
        const readyDeviceId = await Promise.race([
            readyPromise,
            new Promise((resolve, reject) => {
                timeoutId = window.setTimeout(
                    () => reject(new Error('Timed out waiting for Spotify player readiness')),
                    15000,
                );
            }),
        ]).finally(() => window.clearTimeout(timeoutId));

        if (!readyDeviceId || deviceId !== readyDeviceId) {
            throw new Error('Spotify player became unavailable before it was ready');
        }
        return true;
    } catch (error) {
        console.error(error);
        showNotice('Unable to connect the Spotify player. Check your account and browser support, then reload.', true);
        updateControlStates({ playbackEnabled: false });
        return false;
    }
}

async function loadSpotifySDK() {
    if (window.Spotify) {
        return;
    }
    return new Promise((resolve, reject) => {
        const script = document.createElement('script');
        script.src = 'https://sdk.scdn.co/spotify-player.js';
        script.async = true;
        script.onload = () => resolve();
        script.onerror = () => reject(new Error('Failed to load Spotify SDK'));
        document.head.appendChild(script);
    });
}

async function beginScan() {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        showNotice('Camera access is not supported in this browser. Use a modern mobile browser with camera permissions.', true);
        return;
    }

    clearNotice();
    startScanBtn.disabled = true;
    startScanBtn.style.display = 'none';

    try {
        videoStream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } });
        readerDiv.innerHTML = '<video id="webcam-preview" autoplay playsinline muted></video>';
        readerDiv.style.display = 'block';
        // hide buttons while scanning
        playPauseBtn.style.display = 'none';
        restartBtn.style.display = 'none';
        nextBtn.style.display = 'none';
        revealBtn.style.display = 'none';
        const video = document.getElementById('webcam-preview');
        video.srcObject = videoStream;

                if ('BarcodeDetector' in window) {
                    startNativeScanner(video);
                } else {
                    await loadJsQR();
                    startFallbackScanner(video);
                }
    } catch (error) {
        console.error(error);
        showNotice('Camera permission denied or unavailable. Enable camera access and try again.', true);
        startScanBtn.style.display = 'block';
        startScanBtn.disabled = false;
    }
}

async function startScanFromGesture() {
    await activatePlayerFromGesture();
    await beginScan();
}

async function activatePlayerFromGesture() {
    if (player && typeof player.activateElement === 'function') {
        try {
            await player.activateElement();
        } catch (error) {
            console.warn('Could not activate browser audio:', error);
        }
    }
}

function startNativeScanner(video) {
    const detector = new BarcodeDetector({ formats: ['qr_code'] });
    scanInterval = setInterval(async () => {
        try {
            const barcodes = await detector.detect(video);
            if (barcodes.length > 0) {
                stopScan();
                processScanResult(barcodes[0].rawValue);
            }
        } catch (error) {
            // ignore transient camera errors
        }
    }, 150);
}

function startFallbackScanner(video) {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    scanInterval = setInterval(() => {
        if (video.readyState !== video.HAVE_ENOUGH_DATA) {
            return;
        }
        canvas.width = video.videoWidth;
        canvas.height = video.videoHeight;
        ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
        const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
        const code = jsQR(imageData.data, imageData.width, imageData.height, { inversionAttempts: 'dontInvert' });
        if (code) {
            stopScan();
            processScanResult(code.data);
        }
    }, 200);
}

function stopScan() {
    clearInterval(scanInterval);
    scanInterval = null;
    if (videoStream) {
        videoStream.getTracks().forEach((track) => track.stop());
        videoStream = null;
    }
    readerDiv.style.display = 'none';
    readerDiv.innerHTML = '';
}

async function processScanResult(qrCode) {
    clearNotice();
    try {
        const trackData = await fetchJSON('/api/track-info', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ qr_code: qrCode.trim() }),
        });
        currentTrack = trackData;
        updateControlStates({ playbackEnabled: Boolean(deviceId) });
        playbackControls.style.display = 'flex';
        nextControls.style.display = 'grid';
        revealControls.style.display = 'grid';
        playPauseBtn.style.display = 'inline-block';
        restartBtn.style.display = 'inline-block';
        nextBtn.style.display = 'inline-block';
        revealBtn.style.display = 'inline-block';
        revealBtn.textContent = 'Reveal';
        nextBtn.disabled = false;
        const info = document.getElementById('track-info');
        if (info) info.style.display = 'none';
        trackTitle.textContent = 'Hidden until reveal';
        trackArtist.textContent = 'Hidden until reveal';
        trackAlbum.textContent = 'Hidden until reveal';
        trackRelease.textContent = 'Hidden until reveal';
        trackTitle.classList.add('hidden');
        trackArtist.classList.add('hidden');
        trackAlbum.classList.add('hidden');
        trackRelease.classList.add('hidden');
        await activatePlayerFromGesture();
        await playTrack(currentTrack.track_uri);
    } catch (error) {
        console.error(error);
        showNotice('Unable to resolve QR code. Make sure the card contains a valid Spotify track ID or track URL.', true);
        startScanBtn.style.display = 'block';
        startScanBtn.disabled = false;
    }
}

async function playTrack(trackUri, restart = false) {
    if (!deviceId) {
        showNotice('Spotify player is not ready. Reload the page and ensure Spotify Web Playback SDK is supported in this browser.', true);
        return false;
    }
    try {
        await activatePlayerFromGesture();
        const url = `https://api.spotify.com/v1/me/player/play?device_id=${encodeURIComponent(deviceId)}`;
        const body = restart
            ? { uris: [trackUri], position_ms: 0 }
            : { uris: [trackUri] };
        await callSpotifyApi(url, {
            method: 'PUT',
            body: JSON.stringify(body),
        });
        isPlaying = true;
        playPauseBtn.textContent = 'Ⅱ';
        return true;
    } catch (error) {
        console.error(error);
        showNotice(`Playback failed: ${error.message}`, true);
        return false;
    }
}

async function pauseTrack() {
    const url = `https://api.spotify.com/v1/me/player/pause?device_id=${encodeURIComponent(deviceId)}`;
    try {
        await callSpotifyApi(url, { method: 'PUT' });
        isPlaying = false;
        playPauseBtn.textContent = '▶';
    } catch (error) {
        console.error(error);
        showNotice('Unable to pause playback.', true);
    }
}

async function resumeTrack() {
    const url = `https://api.spotify.com/v1/me/player/play?device_id=${encodeURIComponent(deviceId)}`;
    try {
        await activatePlayerFromGesture();
        await callSpotifyApi(url, { method: 'PUT' });
        isPlaying = true;
        playPauseBtn.textContent = 'Ⅱ';
    } catch (error) {
        console.error(error);
        showNotice('Unable to resume playback.', true);
    }
}

async function callSpotifyApi(url, options = {}) {
    if (!accessToken) {
        const tokenData = await fetchJSON('/api/spotify-token');
        accessToken = tokenData.access_token;
    }
    const res = await fetch(url, {
        ...options,
        headers: {
            Authorization: `Bearer ${accessToken}`,
            'Content-Type': 'application/json',
            ...(options.headers || {}),
        },
    });

    if (res.status === 401) {
        await fetchJSON('/api/token-refresh');
        const tokenData = await fetchJSON('/api/spotify-token');
        accessToken = tokenData.access_token;
        const retry = await fetch(url, {
            ...options,
            headers: {
                Authorization: `Bearer ${accessToken}`,
                'Content-Type': 'application/json',
                ...(options.headers || {}),
            },
        });
        if (!retry.ok) {
            throw new Error(await retry.text());
        }
        return retry;
    }
    if (!res.ok) {
        throw new Error(await res.text());
    }
    return res;
}

async function loadJsQR() {
    if (window.jsQR) {
        return;
    }
    return new Promise((resolve, reject) => {
        const script = document.createElement('script');
        script.src = 'https://cdn.jsdelivr.net/npm/jsqr@1.4.0/dist/jsQR.min.js';
        script.async = true;
        script.onload = () => resolve();
        script.onerror = () => reject(new Error('Unable to load jsQR library'));
        document.head.appendChild(script);
    });
}

function resetRound() {
    currentTrack = null;
    isPlaying = false;
    startScanBtn.style.display = 'block';
    startScanBtn.disabled = false;
    playbackControls.style.display = 'none';
    nextControls.style.display = 'none';
    revealControls.style.display = 'none';
    revealBtn.disabled = true;
    nextBtn.disabled = true;
    playPauseBtn.disabled = true;
    restartBtn.disabled = true;
    playPauseBtn.textContent = '▶';
    trackTitle.textContent = 'Hidden until reveal';
    trackArtist.textContent = 'Hidden until reveal';
    trackAlbum.textContent = 'Hidden until reveal';
    trackRelease.textContent = 'Hidden until reveal';
    trackTitle.classList.add('hidden');
    trackArtist.classList.add('hidden');
    trackAlbum.classList.add('hidden');
    trackRelease.classList.add('hidden');
    updateControlStates({ playbackEnabled: false });
    clearNotice();
    // hide track info until next reveal
    revealBtn.style.display = 'inline-block';
    revealBtn.textContent = 'Reveal';
    revealBtn.disabled = true;
    const info = document.getElementById('track-info'); if (info) info.style.display = 'none';
}

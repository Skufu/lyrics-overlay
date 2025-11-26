# SpotLy

<p>
  <a href="https://github.com/Skufu/lyrics-overlay/releases"><img src="https://img.shields.io/github/release/Skufu/lyrics-overlay.svg" alt="Latest Release"></a>
  <a href="https://github.com/Skufu/lyrics-overlay/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/Skufu/lyrics-overlay"><img src="https://goreportcard.com/badge/github.com/Skufu/lyrics-overlay" alt="Go Report Card"></a>
</p>

A transparent, always-on-top lyrics overlay for Spotify. Built for gaming.

<p>
  <img src="https://your-url-here/spotly-demo.gif" width="700" alt="SpotLy running over a game">
</p>

<!-- TODO: Replace with actual demo GIF showing:
     - SpotLy overlay visible during gameplay
     - Time-synced lyrics highlighting in action
     - Transparent background blending with game
-->

SpotLy displays time-synced lyrics in a minimal overlay that stays visible over fullscreen games and streams. No album art, no branding, no clutter. Just the words.


## Demo

<p>
  <img src="https://your-url-here/spotly-gaming.gif" width="700" alt="SpotLy in action while gaming">
</p>

<!-- TODO: Replace with GIF/video showing:
     - Starting SpotLy and connecting to Spotify
     - Overlay appearing over a fullscreen game (VALORANT, League, etc.)
     - Karaoke-style highlighting following the music
     - Adjusting position/opacity mid-game
-->


## Installation

### Download

Grab the latest `spotly.exe` from [Releases](https://github.com/Skufu/lyrics-overlay/releases).

### Build from source

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and build
git clone https://github.com/Skufu/lyrics-overlay.git
cd lyrics-overlay
wails build
```

Requires Go 1.22+ and a [Spotify Developer Account](https://developer.spotify.com/dashboard).


## Tutorial

This is a quick walkthrough on getting SpotLy running.

### Step 1: Create a Spotify App

Go to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard) and create a new app. Under **Redirect URIs**, add:

```
http://127.0.0.1:8080/callback
```

Copy your **Client ID** and **Client Secret**.

### Step 2: Launch SpotLy

Run the executable. On first launch, you'll be prompted to enter your Spotify credentials. Paste in your Client ID and Secret, then click **Connect with Spotify**.

<p>
  <img src="https://your-url-here/spotly-setup.png" width="500" alt="SpotLy setup screen">
</p>

<!-- TODO: Replace with screenshot of the setup/auth screen -->

### Step 3: Authenticate

Your browser will open to Spotify's authorization page. Grant access, and you're done. SpotLy will start displaying lyrics for whatever you're playing.


## Usage

Hover over the overlay to reveal controls:

| Control | Action |
|---------|--------|
| Drag | Reposition the window |
| Lock icon | Prevent accidental movement |
| Settings | Adjust font size, opacity, sync offset |
| Refresh | Force re-fetch current track |

### Sync Offset

If lyrics appear too early or late, adjust the timing slider in settings:

- **Positive values** shift lyrics earlier
- **Negative values** shift lyrics later

Default is 350ms, which works well for most setups.


## Configuration

Config is stored at `~/.spotly/config.json` (Windows: `C:\Users\<YOU>\.spotly\config.json`):

```json
{
  "spotify_client_id": "your_client_id",
  "spotify_client_secret": "your_client_secret",
  "redirect_uri": "http://127.0.0.1:8080/callback",
  "port": 8080,
  "overlay": {
    "x": 100,
    "y": 100,
    "width": 600,
    "height": 120,
    "opacity": 0.9,
    "font_size": 16,
    "visible": true,
    "locked": false,
    "position": "bottom-left",
    "sync_offset": 350
  }
}
```


## Architecture

```
spotly/
├── main.go                 # Wails application entry
├── internal/
│   ├── auth/               # Spotify OAuth2
│   ├── cache/              # LRU lyrics cache
│   ├── config/             # Configuration persistence
│   ├── lyrics/             # LRCLIB provider
│   ├── overlay/            # Display state management
│   └── spotify/            # API client & polling
└── frontend/dist/          # Overlay UI
```

### API Usage

| Service | Endpoint | Purpose |
|---------|----------|---------|
| Spotify | `GET /me/player/currently-playing` | Current track & progress |
| LRCLIB | `GET /api/get` | Synced lyrics lookup |
| LRCLIB | `GET /api/search` | Fallback search |


## Troubleshooting

### OAuth callback fails

- Redirect URI must match exactly: `http://127.0.0.1:8080/callback`
- Ensure port 8080 is available
- Try disabling firewall temporarily

### No lyrics found

- LRCLIB covers most popular songs
- Some tracks don't have lyrics available
- Metadata is normalized automatically

### Overlay not visible in fullscreen

- Use borderless windowed mode
- Some anti-cheat systems block overlays

### Build errors

```bash
wails doctor        # Check dependencies
go mod tidy         # Fix module issues
```


## Roadmap

- [ ] Global hotkeys
- [ ] macOS and Linux support
- [ ] Multi-monitor positioning
- [ ] Custom themes


## Contributing

Contributions welcome. See [contributing guidelines](https://github.com/Skufu/lyrics-overlay/contribute).

Areas of interest:
- Cross-platform support
- Global hotkey implementation
- Performance optimizations


## Acknowledgments

- [Wails](https://wails.io/) - Go desktop framework
- [LRCLIB](https://lrclib.net/) - Synchronized lyrics API
- [Spotify Web API](https://developer.spotify.com/documentation/web-api/)


## License

[MIT](LICENSE)

---

Part of your gaming setup.

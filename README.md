# SpotLy Overlay

A personal-use desktop app that shows time-synced or plain text lyrics in an always-on-top, transparent overlay for the currently playing Spotify track using Go and Wails v2.

## Features

✅ **Implemented:**
- Go 1.22+ with Wails v2 for desktop window/runtime
- Always-on-top, transparent, draggable overlay window
- Spotify OAuth2 Authorization Code flow with localhost callback
- Resilient Spotify API polling with exponential backoff
- LRU cache for lyrics (by track ID and normalized artist|title)
- Lyrics providers: LRCLIB (synced when available) + Genius fallback
- Minimal HTML/CSS/JS frontend for overlay UI
- Configuration persistence in user's home directory

🔄 **In Progress:**
- Dependency resolution and first build
- Global hotkeys implementation
- Enhanced lyrics fetching (full text scraping)

📋 **Planned:**
- Windows packaging and game compatibility testing
- Enhanced UI controls and settings persistence
- Cross-platform support (macOS/Linux)

## Setup Instructions

### Prerequisites

1. **Go 1.22+** installed on your system
2. **Spotify Developer Account** - Register an app at [Spotify Dashboard](https://developer.spotify.com/dashboard)
3. **Genius API Token** (optional) - Get from [Genius API Clients](https://genius.com/api-clients)
   - Not required. LRCLIB is used by default and often returns synced lyrics.

### Installation

1. **Clone and setup:**
   ```bash
   git clone <repository>
   cd lyrics-overlay
   go mod tidy
   ```

2. **Configure Spotify API:**
   - First run will create a config file at `~/.spotly/config.json`.
   - Edit that file and set:
     - `spotify_client_id`
     - `spotify_client_secret`
     - `redirect_uri` should be `http://127.0.0.1:8080/callback` (and must be added in your Spotify Dashboard)
     - Optionally set `genius_token` to enable Genius fallback.
    - Windows path example: `C:\Users\<YOU>\.spotly\config.json`

3. **Install Wails v2 CLI:**
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

4. **Windows quick build:**
   - Double-click `build.bat` or run:
     ```bash
     ./build.bat
     ```

5. **Development run:**
   ```bash
   wails dev
   ```

6. **Build for production:**
   ```bash
   wails build
   ```

### Configuration

The app creates a config file at `~/.spotly/config.json` with these settings:

```json
{
  "spotify_client_id": "your_client_id",
  "spotify_client_secret": "your_client_secret", 
  "redirect_uri": "http://127.0.0.1:8080/callback",
  "port": 8080,
  "genius_token": "optional_genius_token",
  "overlay": {
    "x": 100,
    "y": 100,
    "width": 600,
    "height": 120,
    "opacity": 0.9,
    "font_size": 16,
    "visible": true,
    "locked": false,
    "position": "bottom-left"
  }
}
```

## Usage

1. **First Run:**
   - Launch the app (`spotly.exe` or `wails dev` during development)
   - In the overlay, click "Connect with Spotify"
   - Complete the browser authentication
   - Start playing a song in Spotify
   - The overlay will show the lyrics:
     - LRCLIB first (synced if available),
     - Genius scraped text as fallback (unsynced)

2. **Controls:**
   - Hover the overlay to show controls
   - Toggle visibility, lock position, adjust opacity and font size from the UI

3. **Window Management:**
   - Drag to reposition (when unlocked)
   - Settings persist between sessions
   - Always stays on top of games

## Architecture

```
main.go               # Main application entry (Wails app)
internal/
  ├── auth/          # Spotify OAuth2 flow
  ├── cache/         # LRU lyrics cache
  ├── config/        # Configuration management
  ├── lyrics/        # LyricsProvider interface & implementations
  ├── overlay/       # Window management & display logic
  └── spotify/       # API client & polling service
frontend/dist/       # Wails frontend (HTML/CSS/JS)
```

## Project Status

**Current State:** Core functionality implemented, working on final integration and testing.

**Next Steps:**
1. Complete dependency setup and first successful build
2. Test OAuth flow and Spotify API integration
3. Implement global hotkeys for gaming compatibility
4. Package Windows executable
5. Test overlay behavior in fullscreen games

## API Endpoints Used

- `GET /me/player/currently-playing` - Current track and progress
- `GET /me/player` - Player state (optional, for enhanced device info)
- LRCLIB API: `GET /api/get?track_name=...&artist_name=...` and `GET /api/search?track_name=...&artist_name=...`
- Genius API for search (lyrics scraped from song page as fallback)

## Troubleshooting

### Common Issues

1. **"Could not import github.com/wailsapp/wails/v2"**
   ```bash
   go get github.com/wailsapp/wails/v2@latest
   go mod tidy
   ```

2. **OAuth callback fails:**
   - Ensure redirect URI matches exactly in Spotify Dashboard
   - Check firewall isn't blocking port 8080
   - Try different port in config

3. **Rate limiting (429 errors):**
   - App implements exponential backoff
   - Default polling is 3-5 seconds
   - Increase `POLL_INTERVAL` if needed

4. **No lyrics found:**
   - LRCLIB may not have the track; Genius fallback needs a valid `genius_token`
   - Some tracks may not have lyrics available
   - Ensure artist/title metadata is clean (the app normalizes common suffixes)

5. **Shows translation or wrong text:**
   - LRCLIB should return original lyrics when available
   - Genius fallback filters translation blocks; report specific tracks if you see issues

## Contributing

This is a personal-use project, but contributions are welcome for:
- Additional lyrics providers
- Cross-platform compatibility fixes
- UI/UX improvements
- Performance optimizations

## License

MIT License - see LICENSE file for details.

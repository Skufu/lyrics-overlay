# SpotLy Overlay

<p>

    <a href="https://github.com/Skufu/lyrics-overlay/releases"><img src="https://img.shields.io/github/release/Skufu/lyrics-overlay.svg" alt="Latest Release"></a>

    <a href="https://pkg.go.dev/github.com/Skufu/lyrics-overlay?tab=doc"><img src="https://godoc.org/github.com/Skufu/lyrics-overlay?status.svg" alt="GoDoc"></a>

    <a href="https://github.com/Skufu/lyrics-overlay/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>

</p>

A lightweight, transparent lyrics overlay for Spotify. Built with Go and Wails v2, SpotLy displays time-synced lyrics in an always-on-top window designed for gaming and streaming.


SpotLy is built for gamers and streamers who want lyrics without the clutter. It features time-synced highlighting from LRCLIB with Genius fallback, minimal resource usage, and full customization.

To get started, see the [Quick Start](#quick-start) below, check out the [configuration guide](#configuration), or jump straight to [troubleshooting](#troubleshooting).

## Motivation

I love singing while gaming to vibe with the music, but I'm terrible at memorizing lyrics. As someone who also livestreams, it's embarrassing when you're mid-song and suddenly forget the words.

I couldn't find a lyrics overlay that was truly invisible, lightweight, and game-friendly. Musixmatch has an overlay, but it's a full window with UI, branding, and album art—great for casual use, but terrible for gaming where you need just the lyrics and nothing else.

So I built one.

## Features

- **Always-on-top overlay** – stays visible over fullscreen games

- **Time-synced lyrics** – karaoke-style highlighting that follows the music

- **Transparent design** – adjustable opacity to blend with any game

- **Lightweight** – minimal CPU/GPU usage with smart LRU caching

- **Resilient** – exponential backoff handles Spotify API rate limits gracefully

- **Configurable** – customize position, font size, sync offset, and more

- **Persistent settings** – your preferences are saved between sessions

## Quick Start

### Windows (Recommended)

1. Download the latest `spotly.exe` from [GitHub Releases](https://github.com/Skufu/lyrics-overlay/releases)

2. Run the executable

3. Click "Connect with Spotify" when prompted

4. Complete browser authentication

5. Start playing music in Spotify

### Build from Source

**Prerequisites:**

- Go 1.22+

- Spotify Developer Account ([get credentials](https://developer.spotify.com/dashboard))

- Wails v2 CLI

```

# Install Wails CLI

go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and build

git clone https://github.com/Skufu/lyrics-overlay.git

cd lyrics-overlay

go mod tidy

# Development mode (hot reload)

wails dev

# Production build

wails build

```

**Windows quick build:** Double-click `build.bat` or run `./build.bat`

## Configuration

On first run, a config file is created at `~/.spotly/config.json`:

```

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

**Get your Spotify credentials:**

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)

2. Create a new app

3. Add `http://127.0.0.1:8080/callback` to Redirect URIs

4. Copy Client ID and Secret to your config

**Windows config path:** `C:\Users\<YOU>\.spotly\config.json`

## Usage

### Controls

Hover over the overlay to reveal controls:

- **Drag** – reposition the window (when unlocked)

- **Settings gear** – adjust font size, sync offset, toggle karaoke mode

- **Lock icon** – prevent accidental repositioning while gaming

- **Refresh** – force re-fetch current track and lyrics

### Sync Offset

Adjust the timing slider if lyrics appear too early or too late:

- Positive values → lyrics appear earlier

- Negative values → lyrics appear later

Default offset is 350ms, which works well for most setups.

## Architecture

```

lyrics-overlay/

├── main.go                 # Wails application entry point

├── internal/

│   ├── auth/              # Spotify OAuth2 flow

│   ├── cache/             # LRU lyrics cache

│   ├── config/            # Configuration management

│   ├── lyrics/            # Provider interface (LRCLIB)

│   ├── overlay/           # Window management & display

│   └── spotify/           # API client & polling service

└── frontend/dist/         # HTML/CSS/JS overlay UI

```

### API Endpoints Used

| Service | Endpoint | Purpose |

|---------|----------|---------|

| Spotify | `GET /me/player/currently-playing` | Current track & progress |

| Spotify | `GET /me/player` | Player state |

| LRCLIB | `GET /api/get` | Synced lyrics lookup |

| LRCLIB | `GET /api/search` | Lyrics search fallback |

## Troubleshooting

### OAuth callback fails

- Ensure redirect URI matches exactly in Spotify Dashboard: `http://127.0.0.1:8080/callback`

- Check if port 8080 is available (close other apps using it)

- Temporarily disable firewall and test

- Try changing `port` in config to `8081` or `9000`

### No lyrics found

- LRCLIB is the primary lyrics source (covers most popular songs)

- Some tracks legitimately don't have lyrics available

- App normalizes artist/title metadata automatically

### Rate limiting (429 errors)

- Built-in exponential backoff handles this automatically

- Default polling interval is 4 seconds

- Check `~/.spotly/debug.log` if DEBUG mode is enabled

### Build errors

```

go get github.com/wailsapp/wails/v2@latest

go mod tidy

wails doctor  # Check Wails dependencies

```

### Overlay not visible in fullscreen games

- Try running the game in borderless windowed mode

- Some anti-cheat systems block overlays (expected behavior)

- Check game's overlay compatibility settings

## Roadmap

- [ ] Global hotkeys for hands-free control

- [ ] Cross-platform support (macOS, Linux)

- [ ] Multiple monitor positioning

- [ ] Custom themes and styling


## Contributing

Contributions are welcome! Areas of interest:

- Cross-platform compatibility (macOS/Linux)

- Global hotkey implementation

- UI/UX improvements

- Performance optimizations

- Bug fixes and testing

See [contributing](https://github.com/Skufu/lyrics-overlay/contribute).

## Acknowledgments

SpotLy Overlay is built with the following open source projects:

- [Wails](https://wails.io/) – Go desktop application framework

- [LRCLIB](https://lrclib.net/) – free synchronized lyrics API

- [Spotify Web API](https://developer.spotify.com/documentation/web-api/) – music playback integration

## License

[MIT](https://github.com/Skufu/lyrics-overlay/blob/main/LICENSE)

---

Part of your karaoke gaming setup.


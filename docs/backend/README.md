# Backend Overview

The SpotLy backend is written in Go and organized into modular packages under `internal/`.

## Package Structure

```
internal/
├── auth/       # Spotify OAuth2 authentication
├── cache/      # LRU lyrics caching
├── config/     # Configuration persistence
├── lyrics/     # Lyrics providers (LRCLIB)
├── overlay/    # Display state management
└── spotify/    # Spotify API polling
```

## Entry Point

The main application entry is in [`main.go`](../../main.go), which defines the `App` struct that:
- Initializes all services on startup
- Exposes methods to the frontend via Wails bindings
- Handles application lifecycle

See [main-entry.md](main-entry.md) for details.

---

## Package Documentation

| Package | Purpose | Documentation |
|---------|---------|---------------|
| `auth` | Spotify OAuth2 flow and token management | [auth.md](auth.md) |
| `cache` | LRU cache for lyrics with 24-hour TTL | [cache.md](cache.md) |
| `config` | JSON configuration persistence | [config.md](config.md) |
| `lyrics` | Lyrics fetching from LRCLIB | [lyrics.md](lyrics.md) |
| `overlay` | Display state and timing calculation | [overlay.md](overlay.md) |
| `spotify` | API polling with adaptive intervals | [spotify.md](spotify.md) |

---

## Dependencies

Key external dependencies:

| Package | Purpose |
|---------|---------|
| `github.com/wailsapp/wails/v2` | Desktop application framework |
| `github.com/zmb3/spotify/v2` | Spotify API client |
| `golang.org/x/oauth2` | OAuth2 authentication |
| `golang.org/x/sys/windows` | Windows API calls |

---

## Service Initialization Order

Services are initialized in `OnStartup()` in this order:

1. **Config** → Must be first (other services depend on it)
2. **Cache** → Independent, no dependencies
3. **Overlay** → Depends on Config
4. **Auth** → Depends on Config
5. **Lyrics** → Depends on Cache
6. **Spotify** → Depends on Auth, Overlay, Lyrics

---

## See Also

- [Architecture](../ARCHITECTURE.md) - High-level system overview
- [Data Flow](../DATA_FLOW.md) - How data moves through the system
- [Wails Integration](../wails/README.md) - Frontend bindings

# Main Entry Point

> Documentation for `main.go`, `main_windows.go`, and `main_other.go`

## Overview

The application entry point defines the `App` struct and implements the Wails application lifecycle.

---

## App Struct

```go
type App struct {
    ctx     context.Context
    config  *config.Service
    cache   *cache.Service
    auth    *auth.Service
    overlay *overlay.Service
    spotify *spotify.Service
    lyrics  *lyrics.Service

    // Windows-specific
    overlayHWND      uintptr
    clickThrough     bool
    stopClickMonitor chan struct{}
}
```

---

## Lifecycle Hooks

### `OnStartup(ctx context.Context)`

Called when the app starts. Initializes all services:

1. Config service (loads `~/.spotly/config.json`)
2. Cache service (100-entry LRU)
3. Overlay service (display state)
4. Auth service (Spotify OAuth)
5. Lyrics service (LRCLIB provider)
6. Spotify service (API polling)
7. Click-through monitor (Windows only)

### `OnShutdown(ctx context.Context)`

Called when the app closes:

1. Stops click-through monitor
2. Stops Spotify polling
3. Logs out (clears auth)
4. Shuts down overlay
5. Saves configuration

---

## Exported Methods (Frontend Bindings)

These methods are exposed to JavaScript via Wails:

### Authentication

| Method | Returns | Description |
|--------|---------|-------------|
| `IsAuthenticated()` | `bool` | Check if user is logged in |
| `StartOAuthFlow()` | `error` | Begin Spotify OAuth2 flow |
| `GetAuthURL()` | `(string, error)` | Get OAuth authorization URL |
| `HasCredentials()` | `bool` | Check if Spotify credentials are configured |
| `SaveSpotifyCredentials(id, secret)` | `error` | Save Spotify app credentials |
| `ValidateCredentials(id, secret)` | `error` | Validate credential format |

### Playback & Lyrics

| Method | Returns | Description |
|--------|---------|-------------|
| `GetDisplayInfo()` | `*DisplayInfo` | Current and next lyrics lines |
| `GetSpotifyStatus()` | `map[string]interface{}` | Debug info about connection |
| `TestSpotifyConnection()` | `string` | Test API connectivity |
| `RefreshNow()` | `string` | Force track and lyrics refresh |
| `StartSpotifyPolling()` | `bool` | Manually start polling |

### Overlay Control

| Method | Returns | Description |
|--------|---------|-------------|
| `ToggleVisibility()` | `bool` | Toggle overlay on/off |
| `ResizeWindow(width, height)` | `error` | Resize overlay window |
| `UpdateOverlayConfig(config)` | `error` | Update settings (opacity, font, etc.) |
| `GetOverlayConfig()` | `OverlayConfig` | Get current settings |

### Utilities

| Method | Returns | Description |
|--------|---------|-------------|
| `Quit()` | — | Close the application |
| `GetConfigPath()` | `string` | Path to config file |
| `OpenConfig()` | `(string, error)` | Open config in Explorer |
| `OpenConfigDirectory()` | `error` | Open config folder |

---

## Platform-Specific Code

### Windows (`main_windows.go`)

Implements click-through functionality for gaming:

- `GetActiveWindow()` - Gets foreground window title
- `IsOverlayFocused()` - Checks if SpotLy is focused
- `setOverlayClickThrough(enable)` - Toggles `WS_EX_TRANSPARENT`
- `startClickThroughMonitor()` - Background goroutine checking active window

**Supported games:**
- VALORANT
- League of Legends
- CS2 / Counter-Strike
- Dota 2
- Overwatch
- Apex Legends

### Other Platforms (`main_other.go`)

Stub implementations that do nothing. Click-through is Windows-only.

---

## Wails Configuration

From `wails.json`:

```json
{
  "name": "SpotLy Overlay",
  "outputfilename": "spotly",
  "info": {
    "productName": "SpotLy Overlay",
    "productVersion": "1.0.0"
  }
}
```

---

## See Also

- [Wails Bindings](../wails/bindings.md) - Complete binding reference
- [Architecture](../ARCHITECTURE.md) - System overview
- [Platform Support](../PLATFORM_SUPPORT.md) - Windows features

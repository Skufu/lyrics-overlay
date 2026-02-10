# Config Package

> `internal/config` — Configuration persistence

## Overview

Manages JSON configuration file persistence, including:
- Spotify OAuth credentials
- Overlay display settings
- Authentication tokens

---

## Config File Location

```
~/.spotly/config.json
```

Windows: `C:\Users\<YOU>\.spotly\config.json`

---

## Types

### Config

```go
type Config struct {
    SpotifyClientID     string        `json:"spotify_client_id"`
    SpotifyClientSecret string        `json:"spotify_client_secret"`
    RedirectURI         string        `json:"redirect_uri"`
    Port                int           `json:"port"`
    Overlay             OverlayConfig `json:"overlay"`
    Auth                AuthConfig    `json:"auth"`
}
```

### OverlayConfig

```go
type OverlayConfig struct {
    X            int     `json:"x"`
    Y            int     `json:"y"`
    Width        int     `json:"width"`
    Height       int     `json:"height"`
    Opacity      float64 `json:"opacity"`
    FontSize     int     `json:"font_size"`
    Visible      bool    `json:"visible"`
    Locked       bool    `json:"locked"`
    Position     string  `json:"position"`
    ResizeLocked bool    `json:"resize_locked"`
    SyncOffset   int64   `json:"sync_offset"`
}
```

### AuthConfig

```go
type AuthConfig struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresAt    int64  `json:"expires_at"`
}
```

### Service

```go
type Service struct {
    config   *Config
    filePath string
}
```

---

## Default Values

| Setting | Default |
|---------|---------|
| `port` | `58432` |
| `redirect_uri` | `http://127.0.0.1:58432/callback` |
| `overlay.x` | `100` |
| `overlay.y` | `100` |
| `overlay.width` | `600` |
| `overlay.height` | `120` |
| `overlay.opacity` | `0.9` |
| `overlay.font_size` | `16` |
| `overlay.visible` | `true` |
| `overlay.locked` | `false` |
| `overlay.position` | `"bottom-left"` |
| `overlay.sync_offset` | `350` (ms) |

---

## Key Functions

### Constructor

```go
func New() (*Service, error)
```

1. Creates `~/.spotly/` directory if needed
2. Loads existing config or creates default
3. Returns service instance

### Read/Write

```go
func (s *Service) Get() *Config
func (s *Service) Set(config *Config)
func (s *Service) Load() error
func (s *Service) Save() error
func (s *Service) Path() string
```

### Partial Updates

```go
func (s *Service) UpdateOverlay(overlay OverlayConfig) error
func (s *Service) UpdateAuth(auth AuthConfig) error
```

---

## Example Config File

```json
{
  "spotify_client_id": "your_client_id_here",
  "spotify_client_secret": "your_client_secret_here",
  "redirect_uri": "http://127.0.0.1:58432/callback",
  "port": 58432,
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
    "resize_locked": false,
    "sync_offset": 350
  },
  "auth": {
    "access_token": "...",
    "refresh_token": "...",
    "token_type": "Bearer",
    "expires_at": 1704067200
  }
}
```

---

## See Also

- [Configuration Reference](../CONFIGURATION.md) - User-facing config docs
- [Auth Package](auth.md) - Token management
- [Testing](../TESTING.md) - Config test coverage

# Configuration Reference

Complete reference for SpotLy's configuration file.

---

## Location

```
~/.spotly/config.json
```

| Platform | Path |
|----------|------|
| Windows | `C:\Users\<YOU>\.spotly\config.json` |
| macOS | `/Users/<YOU>/.spotly/config.json` |
| Linux | `/home/<YOU>/.spotly/config.json` |

---

## Full Schema

```json
{
  "spotify_client_id": "string",
  "spotify_client_secret": "string",
  "redirect_uri": "string",
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
    "resize_locked": false,
    "sync_offset": 350
  },
  "auth": {
    "access_token": "string",
    "refresh_token": "string",
    "token_type": "Bearer",
    "expires_at": 1704067200
  }
}
```

---

## Spotify Credentials

| Field | Type | Description |
|-------|------|-------------|
| `spotify_client_id` | string | From Spotify Developer Dashboard |
| `spotify_client_secret` | string | From Spotify Developer Dashboard |
| `redirect_uri` | string | Must match dashboard setting |
| `port` | int | Callback server port |

### Getting Credentials

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create a new app
3. Add redirect URI: `http://127.0.0.1:8080/callback`
4. Copy Client ID and Client Secret

---

## Overlay Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `x` | int | 100 | Window X position |
| `y` | int | 100 | Window Y position |
| `width` | int | 600 | Window width |
| `height` | int | 120 | Window height |
| `opacity` | float | 0.9 | Transparency (0-1) |
| `font_size` | int | 16 | Lyrics font size (px) |
| `visible` | bool | true | Overlay visibility |
| `locked` | bool | false | Prevent dragging |
| `position` | string | "bottom-left" | Reserved for presets |
| `resize_locked` | bool | false | Prevent resizing |
| `sync_offset` | int | 350 | Lyrics timing offset (ms) |

### Sync Offset

- **Positive values** → Lyrics appear earlier
- **Negative values** → Lyrics appear later
- **Range**: -500 to +1000 ms
- **Default**: 350ms (compensates for typical latency)

---

## Auth Tokens

> ⚠️ These are auto-managed. Do not edit manually.

| Field | Type | Description |
|-------|------|-------------|
| `access_token` | string | Spotify API access token |
| `refresh_token` | string | Token for refreshing access |
| `token_type` | string | Always "Bearer" |
| `expires_at` | int | Unix timestamp of expiration |

Tokens are automatically refreshed 5 minutes before expiration.

---

## Example Config

```json
{
  "spotify_client_id": "abc123def456...",
  "spotify_client_secret": "xyz789...",
  "redirect_uri": "http://127.0.0.1:8080/callback",
  "port": 8080,
  "overlay": {
    "x": 50,
    "y": 900,
    "width": 600,
    "height": 120,
    "opacity": 0.85,
    "font_size": 20,
    "visible": true,
    "locked": false,
    "position": "bottom-left",
    "resize_locked": true,
    "sync_offset": 400
  },
  "auth": {
    "access_token": "BQD...",
    "refresh_token": "AQB...",
    "token_type": "Bearer",
    "expires_at": 1704070800
  }
}
```

---

## See Also

- [Development](DEVELOPMENT.md) - First-time setup
- [Config Package](backend/config.md) - Implementation details
- [Troubleshooting](TROUBLESHOOTING.md) - Config issues

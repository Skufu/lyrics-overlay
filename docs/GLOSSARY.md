# Glossary

Terms and concepts used in SpotLy.

---

## A

### Always-on-Top
A window property that keeps the window visible above other windows. SpotLy uses this to stay visible during gameplay.

---

## C

### Cache (LRU)
Least Recently Used cache. Stores recently fetched lyrics in memory. When full, removes oldest entries first. SpotLy uses a 100-entry cache with 24-hour TTL.

### Click-Through
A Windows feature where mouse clicks pass through the window to applications behind it. Enabled automatically when games are detected.

---

## F

### Frameless Window
A window without title bar, borders, or system buttons. SpotLy uses a frameless window for a minimal overlay appearance.

---

## K

### Karaoke Mode
Progressive text highlighting where lyrics change color from left to right as they're sung. Enabled by default in SpotLy.

---

## L

### LRC Format
A timestamped lyrics format used for synchronized lyrics. Example:

```
[00:10.00]First line of the song
[00:15.50]Second line here
```

Format: `[mm:ss.xx]` where mm=minutes, ss=seconds, xx=hundredths.

### LRCLIB
A free, open API for synchronized lyrics at [lrclib.net](https://lrclib.net). SpotLy's primary lyrics source.

---

## O

### OAuth2
An authorization framework that allows third-party applications to access user data without exposing passwords. SpotLy uses OAuth2 to access Spotify's currently-playing track.

### Overlay
A transparent window that displays on top of other applications, typically games.

---

## P

### Polling
Periodically checking an API for updates. SpotLy polls Spotify every 5 seconds when playing music.

---

## S

### Sync Offset
A timing adjustment for lyrics display. Positive values show lyrics earlier, negative values show them later. Default: +350ms.

### Synced Lyrics
Lyrics with timestamps for each line, enabling karaoke-style display. Also called "time-synced" or "LRC" lyrics.

---

## T

### TTL (Time To Live)
Duration before cached data expires. SpotLy uses 24-hour TTL for cached lyrics.

### Token Refresh
OAuth2 mechanism to get a new access token using a refresh token, without requiring user re-authentication.

---

## W

### Wails
A Go framework for building desktop applications with web technologies. SpotLy uses Wails v2 for cross-platform desktop support.

### WebView2
Microsoft's embedded browser component used by Wails on Windows. Uses the Chromium rendering engine.

---

## See Also

- [Architecture](ARCHITECTURE.md) - System overview
- [API Reference](API_REFERENCE.md) - External APIs

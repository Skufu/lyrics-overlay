# Platform Support

Platform-specific features and cross-platform status.

---

## Windows (Primary Platform)

SpotLy is designed primarily for Windows gaming.

### Click-Through Mode

When a supported game is focused, the overlay becomes click-through (mouse events pass to the game).

**How it works:**
1. Background monitor checks active window every 3 seconds
2. If game detected → Enable `WS_EX_TRANSPARENT` style
3. If not in game → Disable click-through (overlay clickable)

**Windows API used:**
- `GetForegroundWindow()` - Get active window handle
- `GetWindowTextW()` - Get window title
- `SetWindowLongW()` - Modify window style

---

### Supported Games

Click-through auto-activates for these games:

| Game | Window Title Match |
|------|-------------------|
| VALORANT | `valorant` |
| League of Legends | `league of legends` |
| CS2 / Counter-Strike | `cs2`, `counter-strike` |
| Dota 2 | `dota 2` |
| Overwatch | `overwatch` |
| Apex Legends | `apex legends` |

Detection is case-insensitive and matches partial window titles.

---

### Code Location

Windows-specific code in `main_windows.go`:

```go
func (a *App) setOverlayClickThrough(enable bool)
func (a *App) startClickThroughMonitor()
func (a *App) GetActiveWindow() (string, error)
```

---

## macOS / Linux

Currently limited support.

### What Works

- Core functionality (OAuth, lyrics, display)
- Overlay rendering via Wails

### What Doesn't Work

- Click-through mode (no implementation)
- Always-on-top may behave differently
- Not tested extensively

### Code Stubs

`main_other.go` provides no-op stubs:

```go
//go:build !windows

func (a *App) setOverlayClickThrough(enable bool) {
    // No-op
}

func (a *App) startClickThroughMonitor() {
    // No-op
}
```

---

## Adding Game Support

To add a new game to click-through detection:

1. Edit `main_windows.go`
2. Add game name to the list:

```go
gamesRequiringClickThrough := []string{
    "valorant",
    "league of legends",
    // Add new game here
    "your game name",
}
```

3. Use lowercase and partial match (e.g., just "fortnite" not full title)

---

## See Also

- [Main Entry](backend/main-entry.md) - Platform code location
- [Troubleshooting](TROUBLESHOOTING.md#overlay-issues) - Overlay issues
- [Development](DEVELOPMENT.md) - Building for different platforms

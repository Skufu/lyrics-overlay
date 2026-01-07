# Wails Bindings Reference

Complete reference of Go methods exposed to the frontend.

---

## Access Pattern

All methods are accessed via:

```javascript
window.go.main.App.MethodName(arguments)
```

All methods return Promises.

---

## Authentication Methods

### `IsAuthenticated() → bool`

Check if user is logged into Spotify.

```javascript
const loggedIn = await window.go.main.App.IsAuthenticated();
```

---

### `StartOAuthFlow() → error`

Begin Spotify OAuth2 flow. Opens browser to Spotify login.

```javascript
await window.go.main.App.StartOAuthFlow();
```

---

### `GetAuthURL() → (string, error)`

Get the OAuth authorization URL for manual flow.

```javascript
const url = await window.go.main.App.GetAuthURL();
```

---

### `HasCredentials() → bool`

Check if Spotify client ID and secret are configured.

```javascript
const configured = await window.go.main.App.HasCredentials();
```

---

### `SaveSpotifyCredentials(clientID, clientSecret) → error`

Save Spotify app credentials to config.

```javascript
await window.go.main.App.SaveSpotifyCredentials(
    "abc123...",
    "xyz789..."
);
```

---

### `ValidateCredentials(clientID, clientSecret) → error`

Validate credential format (minimum length check).

```javascript
await window.go.main.App.ValidateCredentials(id, secret);
```

---

## Playback Methods

### `GetDisplayInfo() → DisplayInfo`

Get current lyrics display information.

```javascript
const info = await window.go.main.App.GetDisplayInfo();
// {
//   current_line: "Current lyrics here",
//   next_line: "Next line preview",
//   is_playing: true,
//   line_duration_ms: 3000,
//   line_progress_ms: 1500,
//   line_start_time_ms: 45000
// }
```

---

### `GetSpotifyStatus() → map[string]interface{}`

Get debug info about Spotify connection.

```javascript
const status = await window.go.main.App.GetSpotifyStatus();
// {
//   authenticated: true,
//   polling: true,
//   has_client: true,
//   current_track: { name: "...", artists: [...], ... }
// }
```

---

### `TestSpotifyConnection() → string`

Test Spotify API connectivity. Returns status emoji + message.

```javascript
const result = await window.go.main.App.TestSpotifyConnection();
// "✅ Found: Song Name by Artist"
// "❌ API Error: ..."
// "⚠️ No active playback"
```

---

### `RefreshNow() → string`

Force immediate track and lyrics refresh.

```javascript
const result = await window.go.main.App.RefreshNow();
```

---

### `StartSpotifyPolling() → bool`

Manually start Spotify polling. Returns true if started.

```javascript
const started = await window.go.main.App.StartSpotifyPolling();
```

---

## Overlay Control Methods

### `ToggleVisibility() → bool`

Toggle overlay visibility. Returns new state.

```javascript
const visible = await window.go.main.App.ToggleVisibility();
```

---

### `ResizeWindow(width, height) → error`

Resize the overlay window.

```javascript
await window.go.main.App.ResizeWindow(600, 120);
```

---

### `UpdateOverlayConfig(config) → error`

Update overlay settings. Pass partial object with fields to update.

```javascript
await window.go.main.App.UpdateOverlayConfig({
    opacity: 0.9,
    font_size: 18,
    sync_offset: 350,
    locked: false,
    visible: true
});
```

---

### `GetOverlayConfig() → OverlayConfig`

Get current overlay configuration.

```javascript
const config = await window.go.main.App.GetOverlayConfig();
// {
//   x: 100, y: 100,
//   width: 600, height: 120,
//   opacity: 0.9,
//   font_size: 16,
//   visible: true,
//   locked: false,
//   position: "bottom-left",
//   resize_locked: false,
//   sync_offset: 350
// }
```

---

## Utility Methods

### `Quit()`

Close the application.

```javascript
window.go.main.App.Quit();
```

---

### `GetConfigPath() → string`

Get full path to config file.

```javascript
const path = await window.go.main.App.GetConfigPath();
// "C:\\Users\\User\\.spotly\\config.json"
```

---

### `OpenConfig() → (string, error)`

Open config file location in Explorer. Returns path.

```javascript
const path = await window.go.main.App.OpenConfig();
```

---

### `OpenConfigDirectory() → error`

Open config folder in file explorer.

```javascript
await window.go.main.App.OpenConfigDirectory();
```

---

## Wails Runtime

Additional runtime functions (not Go bindings):

```javascript
// Close application
window.runtime.Quit();

// Window position
const pos = await window.runtime.WindowGetPosition();
// { x: 100, y: 200 }

window.runtime.WindowSetPosition(150, 250);

// Window size
const size = await window.runtime.WindowGetSize();
// { w: 600, h: 120 }

window.runtime.WindowSetSize(700, 150);
```

---

## See Also

- [Main Entry](../backend/main-entry.md) - Method implementations
- [Wails Integration](README.md) - Framework overview
- [Frontend UI](../frontend/ui-components.md) - Usage examples

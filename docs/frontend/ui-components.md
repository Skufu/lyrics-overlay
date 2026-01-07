# UI Components

Detailed documentation of frontend UI elements and JavaScript logic.

---

## Auth Modal

Shown when user is not authenticated.

### Elements

| ID | Purpose |
|----|---------|
| `auth-screen` | Modal container |
| `auth-btn` | "Connect with Spotify" button |
| `auth-status` | Status message display |
| `auth-setup-hint` | Links to setup wizard and config |

### JavaScript Functions

```javascript
async function startAuth()
```
Initiates OAuth flow if credentials exist, else prompts for setup.

```javascript
function checkAuthComplete()
```
Polls `IsAuthenticated()` every second until user completes OAuth.

---

## Setup Wizard

First-time configuration for Spotify credentials.

### Elements

| ID | Purpose |
|----|---------|
| `setup-wizard` | Modal container |
| `setup-client-id` | Client ID input |
| `setup-client-secret` | Client Secret input |
| `setup-save-btn` | Save credentials button |
| `setup-error` | Error message display |
| `setup-success` | Success message display |

### JavaScript Functions

```javascript
async function showSetupWizard()
async function closeSetupWizard()
function copyRedirectUri(element)
async function saveSetupCredentials()
```

---

## Main Overlay

Lyrics display after authentication.

### Elements

| ID | Purpose |
|----|---------|
| `main-overlay` | Container (displayed when authenticated) |
| `current-line` | Current lyrics line |
| `next-line` | Next lyrics preview |
| `statusIndicator` | Green/red playing indicator |
| `settingsToggle` | Gear icon to open settings |

### JavaScript Functions

```javascript
function startDisplayInfoPolling()
```
Polls `GetDisplayInfo()` every 50ms for smooth karaoke effect.

```javascript
function showSettingsGear(e)
function hideSettingsGear()
```
Toggle visibility of settings gear on lyrics click.

---

## Settings Panel

Expandable configuration options.

### Sections

1. **Track Info** - Current song name and artist
2. **Appearance** - Font size slider
3. **Sync Timing** - Offset slider (-500ms to +1000ms)
4. **Effects** - Karaoke mode toggle
5. **Controls** - Refresh, Lock, Logout, Quit
6. **Help** - Setup wizard, Config file access

### JavaScript Functions

```javascript
async function toggleSettings()
function updateFontSize(value)
function updateSyncOffset(value)
async function toggleLock()
function toggleKaraoke()
async function refreshNow()
async function logout()
```

---

## Window Dragging

Custom drag implementation for frameless window.

```javascript
let dragging = false;
let dragStart = { x: 0, y: 0 };
let winStart = { x: 0, y: 0 };

async function onDragStart(e)
function onDragMove(e)
function onDragEnd()
```

Uses Wails runtime:
- `window.runtime.WindowGetPosition()`
- `window.runtime.WindowSetPosition(x, y)`

---

## Karaoke Animation

### State Variables

```javascript
let karaokeEnabled = true;
let lastLineStartTime = 0;
```

### Update Loop

Inside `startDisplayInfoPolling()`:

```javascript
if (karaokeEnabled && info.is_playing && info.line_duration_ms > 0) {
    const progress = (info.line_progress_ms / info.line_duration_ms) * 100;
    currentEl.style.setProperty('--karaoke-progress', progress + '%');
}
```

---

## Lyrics Display Animation

### CSS Animations

```css
@keyframes slideUp {
    from { transform: translateY(6px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
}

.lyrics-display.updating .current-line {
    animation: slideUp 0.2s ease-out;
}
```

### Trigger

```javascript
if (lastCurrentLine !== newCurrent) {
    lyricsDisplay.classList.add('updating');
    // ... update content
    setTimeout(() => lyricsDisplay.classList.remove('updating'), 200);
}
```

---

## Backend Communication

All Go methods accessed via `window.go.main.App.*`:

```javascript
// Authentication
await window.go.main.App.IsAuthenticated();
await window.go.main.App.StartOAuthFlow();
await window.go.main.App.HasCredentials();
await window.go.main.App.SaveSpotifyCredentials(id, secret);

// Playback
await window.go.main.App.GetDisplayInfo();
await window.go.main.App.GetSpotifyStatus();
await window.go.main.App.RefreshNow();

// Settings
await window.go.main.App.UpdateOverlayConfig({ opacity: 0.9 });
await window.go.main.App.GetOverlayConfig();
await window.go.main.App.ResizeWindow(600, 120);

// Wails runtime
window.runtime.Quit();
window.runtime.WindowGetPosition();
window.runtime.WindowSetPosition(x, y);
```

---

## See Also

- [Wails Bindings](../wails/bindings.md) - Complete method reference
- [Overlay Package](../backend/overlay.md) - DisplayInfo structure
- [Frontend README](README.md) - Overview and structure

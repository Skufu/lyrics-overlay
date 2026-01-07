# Wails Integration

SpotLy uses [Wails v2](https://wails.io/) as the desktop application framework.

## Overview

Wails enables:
- **Go backend** with full standard library access
- **Web frontend** using HTML/CSS/JS
- **Native bindings** - Go functions callable from JavaScript
- **Cross-platform** - Windows, macOS, Linux (with limitations)

---

## Project Configuration

From `wails.json`:

```json
{
  "name": "SpotLy Overlay",
  "outputfilename": "spotly",
  "frontend:install": "",
  "frontend:build": "",
  "info": {
    "productName": "SpotLy Overlay",
    "productVersion": "1.0.0"
  }
}
```

**Note:** Frontend install/build commands are empty because the frontend is pre-built static HTML.

---

## Application Options

From `main.go`:

```go
wails.Run(&options.App{
    Title:            "SpotLy Overlay",
    Width:            600,
    Height:           500,
    Frameless:        true,
    AlwaysOnTop:      true,
    BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
    DisableResize:    disableResizeAtStartup,
    Windows: &wailswindows.Options{
        WebviewIsTransparent:              true,
        WindowIsTranslucent:               true,
        DisableFramelessWindowDecorations: true,
    },
    AssetServer: &assetserver.Options{
        Assets: assets,  // Embedded frontend/dist
    },
    OnStartup:  app.OnStartup,
    OnShutdown: app.OnShutdown,
    Bind:       []interface{}{app},
})
```

### Key Options

| Option | Value | Purpose |
|--------|-------|---------|
| `Frameless` | `true` | No title bar or borders |
| `AlwaysOnTop` | `true` | Stays above other windows |
| `BackgroundColour` | Transparent | See-through background |
| `WebviewIsTransparent` | `true` | WebView2 transparency (Windows) |
| `WindowIsTranslucent` | `true` | Window transparency (Windows) |

---

## Asset Embedding

```go
//go:embed all:frontend/dist
var assets embed.FS
```

The entire `frontend/dist` directory is embedded into the binary at compile time.

---

## Binding System

The `App` struct is bound via:

```go
Bind: []interface{}{app}
```

All exported methods on `App` become available in JavaScript as:

```javascript
window.go.main.App.MethodName(args)
```

---

## Runtime APIs

Wails provides runtime functions via `window.runtime`:

| Function | Purpose |
|----------|---------|
| `Quit()` | Close application |
| `WindowGetPosition()` | Get window X, Y |
| `WindowSetPosition(x, y)` | Move window |
| `WindowGetSize()` | Get window dimensions |
| `WindowSetSize(w, h)` | Resize window |

---

## Build Commands

```bash
# Development (with hot reload)
wails dev

# Production build
wails build

# Check dependencies
wails doctor
```

---

## See Also

- [Bindings Reference](bindings.md) - Complete method list
- [Main Entry](../backend/main-entry.md) - App struct and lifecycle
- [Development Guide](../DEVELOPMENT.md) - Build instructions

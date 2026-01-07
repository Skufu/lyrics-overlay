# Frontend Overview

The SpotLy frontend is a single-page HTML/JS/CSS application that serves as the overlay UI.

## Structure

```
frontend/
├── dist/
│   └── index.html    # Complete application (HTML + CSS + JS)
└── wailsjs/
    ├── go/           # Generated Go bindings
    └── runtime/      # Wails runtime library
```

## Technology

- **Pure HTML/CSS/JS** - No framework (React, Vue, etc.)
- **Embedded at build** - `//go:embed all:frontend/dist` in `main.go`
- **Wails bindings** - Auto-generated TypeScript/JS for Go methods

---

## Key Features

### Transparent Window

```css
html, body, :root {
    background: transparent !important;
}
```

The window is frameless with transparent background, allowing overlay behavior.

### CSS Variables

```css
:root {
    --spotify-green: #1db954;
    --spotify-green-light: #1ed760;
    --bg-dark: rgba(0, 0, 0, 0.85);
    --bg-panel: rgba(0, 0, 0, 0.72);
    --text-primary: #ffffff;
    --text-secondary: rgba(255, 255, 255, 0.7);
    --text-muted: rgba(255, 255, 255, 0.5);
}
```

### Karaoke Effect

Progressive text highlighting using CSS gradient:

```css
.current-line.karaoke {
    background: linear-gradient(
        90deg,
        var(--spotify-green) var(--karaoke-progress, 0%),
        #ffffff var(--karaoke-progress, 0%)
    );
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
}
```

---

## UI States

| State | Description |
|-------|-------------|
| **Auth Modal** | Shown when not authenticated |
| **Setup Wizard** | First-time credential configuration |
| **Main Overlay** | Lyrics display after authentication |
| **Settings Panel** | Expandable configuration options |

---

## Window Dimensions

| State | Width | Height |
|-------|-------|--------|
| Auth Screen | 600px | 600px |
| Setup Wizard | 600px | 580px |
| Compact (lyrics) | 600px | 120px |
| Expanded (settings) | 600px | 560px |

---

## See Also

- [UI Components](ui-components.md) - Detailed component documentation
- [Wails Bindings](../wails/bindings.md) - JavaScript API access
- [Architecture](../ARCHITECTURE.md) - System overview

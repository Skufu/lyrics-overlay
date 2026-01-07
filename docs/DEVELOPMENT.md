# Development Guide

Getting started with SpotLy development.

---

## Prerequisites

| Requirement | Version | Installation |
|-------------|---------|--------------|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl/) |
| Wails CLI | Latest | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Spotify Account | — | [developer.spotify.com](https://developer.spotify.com) |

### Windows-Specific

- WebView2 Runtime (usually pre-installed on Windows 10/11)

---

## Quick Start

```bash
# Clone repository
git clone https://github.com/Skufu/lyrics-overlay.git
cd lyrics-overlay

# Check dependencies
wails doctor

# Run in development mode
wails dev

# Build production binary
wails build
```

---

## Spotify App Setup

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Click "Create App"
3. Fill in app name and description
4. Add Redirect URI: `http://127.0.0.1:8080/callback`
5. Copy **Client ID** and **Client Secret**

---

## Running

### Development Mode

```bash
wails dev
```

Features:
- Hot reload on Go changes
- Frontend changes require manual refresh
- Opens DevTools automatically

### Production Build

```bash
wails build
```

Output: `build/bin/spotly.exe`

---

## Testing

```bash
# Run all tests
go test -v ./internal/...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./internal/...

# View coverage report
go tool cover -html=coverage.out
```

---

## Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

Configuration in `.golangci.yml`.

---

## Project Structure

```
lyrics-overlay/
├── main.go              # Wails entry point, App struct
├── main_windows.go      # Windows-specific code
├── main_other.go        # Non-Windows stubs
├── wails.json           # Wails configuration
├── go.mod               # Go module
├── internal/            # Backend packages
│   ├── auth/            # Spotify OAuth
│   ├── cache/           # LRU cache
│   ├── config/          # Configuration
│   ├── lyrics/          # LRCLIB provider
│   ├── overlay/         # Display state
│   └── spotify/         # API polling
├── frontend/
│   └── dist/            # Frontend assets
│       └── index.html   # Complete UI
└── docs/                # Documentation
```

---

## CI/CD

### CI Pipeline (`.github/workflows/ci.yml`)

Runs on push/PR to main:
1. **Test job**: `go test -v -race ./internal/...`
2. **Lint job**: `golangci-lint`
3. **Build job**: `wails build` (Windows)

### Release Pipeline (`.github/workflows/release.yml`)

Triggered by version tags (`v*`):
1. Build Windows executable
2. Create GitHub release
3. Upload artifacts

---

## Making Changes

### Backend

1. Modify Go files in `internal/` or `main*.go`
2. Run `wails dev` to test
3. Add tests for new functionality
4. Run `golangci-lint run` before committing

### Frontend

1. Edit `frontend/dist/index.html`
2. Refresh browser in dev mode
3. Changes are embedded in binary on build

---

## See Also

- [Architecture](ARCHITECTURE.md) - System overview
- [Testing](TESTING.md) - Test coverage details
- [Backend Overview](backend/README.md) - Package docs

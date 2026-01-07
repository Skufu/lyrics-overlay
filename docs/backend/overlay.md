# Overlay Package

> `internal/overlay` — Display state management

## Overview

Manages the overlay display state, including:
- Current track information
- Current lyrics data
- Display info calculation (current line, next line, timing)
- Visibility and configuration

---

## Types

### Service

```go
type Service struct {
    config        *config.Service
    mu            sync.RWMutex
    currentTrack  *TrackInfo
    currentLyrics *LyricsData
    isVisible     bool
    lastUpdate    time.Time
}
```

### TrackInfo

```go
type TrackInfo struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Artists   []string  `json:"artists"`
    Album     string    `json:"album"`
    Duration  int64     `json:"duration_ms"`
    Progress  int64     `json:"progress_ms"`
    IsPlaying bool      `json:"is_playing"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### LyricsData

```go
type LyricsData struct {
    TrackID   string       `json:"track_id"`
    Source    string       `json:"source"`
    Lines     []LyricsLine `json:"lines"`
    IsSynced  bool         `json:"is_synced"`
    FetchedAt time.Time    `json:"fetched_at"`
}
```

### LyricsLine

```go
type LyricsLine struct {
    Text      string `json:"text"`
    Timestamp int64  `json:"timestamp_ms,omitempty"`
}
```

### DisplayInfo

```go
type DisplayInfo struct {
    CurrentLine   string `json:"current_line"`
    NextLine      string `json:"next_line"`
    IsPlaying     bool   `json:"is_playing"`
    LineDuration  int64  `json:"line_duration_ms"`
    LineProgress  int64  `json:"line_progress_ms"`
    LineStartTime int64  `json:"line_start_time_ms"`
}
```

---

## Key Functions

### Track/Lyrics Management

```go
func (s *Service) GetCurrentTrack() *TrackInfo
func (s *Service) SetCurrentTrack(track *TrackInfo)
func (s *Service) GetCurrentLyrics() *LyricsData
func (s *Service) SetCurrentLyrics(lyrics *LyricsData)
```

### Display Calculation

```go
func (s *Service) GetDisplayInfo() *DisplayInfo
```

**Algorithm for synced lyrics:**

1. Get current playback progress (stored + elapsed time)
2. Apply sync offset (default: 350ms ahead)
3. Find lyrics line where `timestamp <= progress`
4. Calculate line duration and progress for karaoke effect
5. Skip empty lines for current display
6. Return current line, next line, and timing info

### Visibility Control

```go
func (s *Service) ToggleVisibility() bool
func (s *Service) IsVisible() bool
func (s *Service) SetVisibility(visible bool)
```

### Configuration

```go
func (s *Service) GetOverlayConfig() config.OverlayConfig
func (s *Service) UpdateOverlayConfig(overlayConfig config.OverlayConfig) error
```

---

## Sync Offset

The `SyncOffset` (default 350ms) shifts lyrics timing:

- **Positive value** → Lyrics appear earlier
- **Negative value** → Lyrics appear later

This compensates for:
- Spotify API latency
- Audio processing delay
- Personal preference

---

## Progress Calculation

For smooth karaoke highlighting, the overlay calculates real-time progress:

```go
// Derive effective progress
progress := track.Progress
if track.IsPlaying {
    elapsed := time.Since(track.UpdatedAt).Milliseconds()
    progress += elapsed
}
progress += syncOffset
```

This allows 50ms UI updates without waiting for API polls.

---

## State Diagram

```
┌─────────────────────────────────────────────┐
│              Overlay Service                │
├─────────────────────────────────────────────┤
│  currentTrack   ◄──── SetCurrentTrack()     │
│  currentLyrics  ◄──── SetCurrentLyrics()    │
│       │               │                     │
│       ▼               ▼                     │
│  ┌─────────────────────────┐               │
│  │   GetDisplayInfo()      │               │
│  │   - Calculate progress  │               │
│  │   - Find current line   │               │
│  │   - Calculate timing    │               │
│  └───────────┬─────────────┘               │
│              ▼                              │
│       DisplayInfo                           │
│       (sent to frontend)                    │
└─────────────────────────────────────────────┘
```

---

## See Also

- [Spotify Package](spotify.md) - Sets track data
- [Lyrics Package](lyrics.md) - Sets lyrics data
- [Frontend UI](../frontend/ui-components.md) - Displays the info

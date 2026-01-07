# Spotify Package

> `internal/spotify` — Spotify API polling

## Overview

Handles Spotify API interactions with:
- Background polling loop
- Adaptive interval adjustment
- Rate limit handling
- Track change detection

---

## Types

### Service

```go
type Service struct {
    auth              *auth.Service
    overlay           *overlay.Service
    lyrics            *lyrics.Service
    stopChan          chan struct{}
    isPolling         bool
    baseInterval      time.Duration  // 5 seconds
    currentInterval   time.Duration
    backoffFactor     float64        // 1.5x
    maxInterval       time.Duration  // 30 seconds
    lastTrackID       string
    consecutiveErrors int
}
```

---

## Key Functions

### Constructor

```go
func New(authSvc *auth.Service, overlaySvc *overlay.Service, lyricsSvc *lyrics.Service) *Service
```

### Lifecycle

```go
func (s *Service) Start()
func (s *Service) Stop()
func (s *Service) IsPolling() bool
```

### Manual Access

```go
func (s *Service) GetCurrentTrack() *overlay.TrackInfo
```

---

## Polling Loop

```go
func (s *Service) pollLoop()
```

Runs in a goroutine, polling Spotify API at `currentInterval`.

### Poll Flow

```
┌─────────────────┐
│ pollLoop()      │
└────────┬────────┘
         │
         ▼
┌────────────────────────┐
│ auth.GetClient()       │ ─── nil ──► Set no track, backoff
└────────┬───────────────┘
         │
         ▼
┌────────────────────────┐
│ client.PlayerPlaying() │ ─── error ─► handleError()
└────────┬───────────────┘
         │
         ▼
┌────────────────────────┐
│ Check for playback     │ ─── nil ──► handleNoPlayback()
└────────┬───────────────┘
         │
         ▼
┌────────────────────────┐
│ extractTrackInfo()     │
└────────┬───────────────┘
         │
         ▼
┌────────────────────────┐
│ Track changed?         │ ─── yes ──► fetchAndSetLyrics()
└────────┬───────────────┘
         │
         ▼
┌────────────────────────┐
│ overlay.SetTrack()     │
└────────┬───────────────┘
         │
         ▼
┌────────────────────────┐
│ adjustInterval()       │
└────────────────────────┘
```

---

## Adaptive Intervals

| State | Interval |
|-------|----------|
| Music playing | 5 seconds (base) |
| Paused/No content | 15 seconds (base × 3) |
| Errors | Exponential backoff (×1.5, max 30s) |
| Rate limited (429) | 30 seconds (max) |

```go
func (s *Service) adjustInterval(isPlaying, hasError bool)
```

---

## Track Change Detection

On track change:
1. Reset interval to base
2. Clear error count
3. Trigger async lyrics fetch

```go
if track.ID != s.lastTrackID {
    s.lastTrackID = track.ID
    s.resetInterval()
    go s.fetchAndSetLyrics(track)
}
```

---

## Error Handling

### Rate Limiting (429)

```go
func (s *Service) handleRateLimit(err *spotify.Error)
```

Sets interval to maximum (30 seconds).

### General Errors

```go
func (s *Service) handleError(err error)
```

- After 3 consecutive errors: Apply backoff
- After 5 consecutive errors: Clear current track

### No Playback

```go
func (s *Service) handleNoPlayback()
```

Clears track and applies backoff.

---

## Extract Track Info

```go
func (s *Service) extractTrackInfo(playerState *spotify.CurrentlyPlaying) *overlay.TrackInfo
```

Maps Spotify API response to internal `TrackInfo`:

| API Field | Internal Field |
|-----------|----------------|
| `Item.ID` | `ID` |
| `Item.Name` | `Name` |
| `Item.Artists[].Name` | `Artists` |
| `Item.Album.Name` | `Album` |
| `Item.Duration` | `Duration` |
| `Progress` | `Progress` |
| `Playing` | `IsPlaying` |
| — | `UpdatedAt` (current time) |

---

## Dependencies

```
┌─────────────┐
│   Spotify   │
│   Service   │
└──────┬──────┘
       │
       ├──────► auth.Service (GetClient)
       │
       ├──────► overlay.Service (SetCurrentTrack, SetCurrentLyrics)
       │
       └──────► lyrics.Service (GetLyrics)
```

---

## See Also

- [Auth Package](auth.md) - Provides authenticated client
- [Overlay Package](overlay.md) - Receives track/lyrics updates
- [Lyrics Package](lyrics.md) - Fetches lyrics on track change
- [API Reference](../API_REFERENCE.md#spotify) - Spotify endpoints

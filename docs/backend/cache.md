# Cache Package

> `internal/cache` — LRU lyrics caching

## Overview

Implements an in-memory LRU (Least Recently Used) cache for lyrics data with:
- Dual-key lookup (track ID and normalized artist|title)
- 24-hour TTL (time-to-live)
- Configurable max size with automatic eviction

---

## Types

### Service

```go
type Service struct {
    mu          sync.RWMutex
    maxSize     int
    trackCache  map[string]*cacheEntry   // By Spotify track ID
    keyCache    map[string]*cacheEntry   // By normalized "artist|title"
    lruList     *list.List               // For eviction ordering
    trackToElem map[string]*list.Element
    keyToElem   map[string]*list.Element
}
```

### CacheStats

```go
type CacheStats struct {
    Size         int `json:"size"`
    MaxSize      int `json:"max_size"`
    TrackEntries int `json:"track_entries"`
    KeyEntries   int `json:"key_entries"`
}
```

---

## Key Functions

### Constructor

```go
func New(maxSize int) *Service
```

Creates cache with specified max size. Default: 100 entries.

---

### Get Operations

```go
func (s *Service) GetByTrackID(trackID string) *overlay.LyricsData
```

Lookup by Spotify track ID. Returns `nil` if not found or expired (>24 hours).

```go
func (s *Service) GetByKey(cacheKey string) *overlay.LyricsData
```

Lookup by normalized cache key (format: `"artist|title"`).

---

### Set Operations

```go
func (s *Service) SetByTrackID(trackID string, lyrics *overlay.LyricsData)
```

Cache lyrics by Spotify track ID.

```go
func (s *Service) SetByKey(cacheKey string, lyrics *overlay.LyricsData)
```

Cache lyrics by normalized key.

---

### Maintenance

```go
func (s *Service) Clear()
```

Remove all cache entries.

```go
func (s *Service) Size() int
```

Current number of cached entries.

```go
func (s *Service) Stats() CacheStats
```

Cache statistics for debugging.

---

## How It Works

### Dual-Key System

The same lyrics can be looked up by:
1. **Track ID** - Exact Spotify track identifier
2. **Cache Key** - Normalized `"artist|title"` string

This helps when the same song appears with different track IDs (e.g., remastered versions).

### LRU Eviction

When cache exceeds `maxSize`:
1. Find least recently used entry (back of list)
2. Remove from all maps
3. Repeat until under limit

### TTL Expiration

Entries older than 24 hours are considered stale and removed on access.

---

## Cache Flow

```
┌─────────────┐
│ GetLyrics() │
└──────┬──────┘
       │
       ▼
┌──────────────────┐     Hit
│ GetByTrackID()   │────────────► Return lyrics
└────────┬─────────┘
         │ Miss
         ▼
┌──────────────────┐     Hit
│ GetByKey()       │────────────► Return lyrics
└────────┬─────────┘              + cache by track ID
         │ Miss
         ▼
┌──────────────────┐
│ Fetch from       │
│ LRCLIB API       │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ SetByTrackID()   │
│ SetByKey()       │
└──────────────────┘
```

---

## Thread Safety

All operations use `sync.RWMutex`:
- **Read operations** (`Get*`, `Size`, `Stats`) - `RLock`
- **Write operations** (`Set*`, `Clear`) - `Lock`

---

## See Also

- [Lyrics Package](lyrics.md) - Uses cache for lookup
- [Testing](../TESTING.md) - Cache test coverage

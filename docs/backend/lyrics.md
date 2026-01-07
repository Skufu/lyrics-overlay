# Lyrics Package

> `internal/lyrics` — Lyrics providers and LRC parsing

## Overview

Fetches synchronized lyrics from external providers, with:
- Provider interface for extensibility
- LRCLIB as primary source
- LRC format parsing
- Title/artist normalization for better matching
- Demo fallback for missing lyrics

---

## Types

### LyricsProvider Interface

```go
type LyricsProvider interface {
    SearchLyrics(artist, title string) (*overlay.LyricsData, error)
    GetName() string
}
```

### Service

```go
type Service struct {
    providers []LyricsProvider
    cache     *cache.Service
    client    *http.Client
}
```

---

## Providers

### LRCLibProvider

Primary lyrics source. Queries [lrclib.net](https://lrclib.net) API.

```go
type LRCLibProvider struct {
    client  *http.Client
    baseURL string  // https://lrclib.net/api
}
```

**Endpoints used:**
- `GET /api/get?track_name=...&artist_name=...` - Exact match
- `GET /api/search?track_name=...&artist_name=...` - Fuzzy search
- `GET /api/search?q=...` - Query fallback
- `GET /api/get/{id}` - Fetch by ID

### DemoProvider

Fallback when no lyrics found. Returns basic track info display.

---

## Key Functions

### Constructor

```go
func New(cacheSvc *cache.Service) *Service
```

Creates service with LRCLIB and Demo providers.

### Main Lookup

```go
func (s *Service) GetLyrics(trackID, artist, title string) (*overlay.LyricsData, error)
```

1. Check cache by track ID
2. Check cache by normalized key
3. Try each provider in order
4. Cache successful results
5. Return lyrics or error

---

## LRC Parsing

```go
func parseLRCToLines(lrc string) []overlay.LyricsLine
```

Parses LRC format timestamps:

```
[00:12.34]First line
[00:15.67]Second line
```

**Handles:**
- `[mm:ss.xx]` and `[mm:ss.xxx]` formats
- Multiple timestamps per line
- Metadata tags (`[ti:]`, `[ar:]`, etc.) - skipped
- Unsorted input - auto-sorted by timestamp

---

## Title Normalization

```go
func normalizeString(text string) string
```

Removes common patterns for better matching:
- `(feat. ...)`, `(ft. ...)`, `(featuring ...)`
- `[Remastered]`, `[anything in brackets]`
- `(Remix)`, `(Version)`, `(Edit)`
- `- Remaster`, `- Radio Edit`
- Special characters and extra whitespace

**Examples:**
| Input | Normalized |
|-------|------------|
| `Song (feat. Artist)` | `song` |
| `Track [Remastered 2024]` | `track` |
| `Title - Radio Edit` | `title` |

---

## Search Strategy

```
┌─────────────────┐
│ tryGet()        │ ◄── Exact match by artist/title
└────────┬────────┘
         │ Miss
         ▼
┌─────────────────┐
│ search()        │ ◄── Fuzzy search by artist/title
└────────┬────────┘
         │ Empty
         ▼
┌─────────────────┐
│ searchByQuery() │ ◄── Combined query "title artist"
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ pickBestMatch() │ ◄── Score results, pick best
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ getByID()       │ ◄── Fetch full lyrics for best match
└─────────────────┘
```

---

## Scoring Algorithm

```go
func pickBestLRCLibMatch(results []lrcLibTrack, artist, title string) *lrcLibTrack
```

| Criterion | Score |
|-----------|-------|
| Exact artist match | +3 |
| Exact title match | +3 |
| Has synced lyrics | +2 |
| Has plain lyrics | +1 |

---

## Noise Filtering

Plain lyrics text is filtered to remove:
- "You might also like"
- "Genius annotation"
- Contributor/translation UI strings
- Language names (Deutsch, Español, etc.)
- Embed counts (e.g., "123Embed")

---

## See Also

- [Cache Package](cache.md) - Lyrics caching
- [API Reference](../API_REFERENCE.md#lrclib) - LRCLIB endpoints
- [Glossary](../GLOSSARY.md#lrc-format) - LRC format explanation

# API Reference

External APIs used by SpotLy.

---

## Spotify Web API

Base URL: `https://api.spotify.com/v1`

### Authentication

SpotLy uses OAuth2 Authorization Code flow with these scopes:
- `user-read-currently-playing`
- `user-read-playback-state`

---

### GET /me/player/currently-playing

Returns the currently playing track.

**Headers:**
```
Authorization: Bearer {access_token}
```

**Response (200 OK):**
```json
{
  "item": {
    "id": "spotify:track:abc123",
    "name": "Song Name",
    "artists": [
      { "name": "Artist Name" }
    ],
    "album": {
      "name": "Album Name"
    },
    "duration_ms": 240000
  },
  "progress_ms": 45000,
  "is_playing": true
}
```

**Used by:** `internal/spotify` polling loop

**Rate Limits:** No hard limit, but respect fair use (~1 req/sec)

---

## LRCLIB API

Base URL: `https://lrclib.net/api`

No authentication required.

---

### GET /api/get

Exact match lookup by track and artist.

**Parameters:**
| Name | Required | Description |
|------|----------|-------------|
| `track_name` | Yes | Song title |
| `artist_name` | Yes | Artist name |
| `album_name` | No | Album name |
| `duration` | No | Track duration (seconds) |

**Example:**
```
GET /api/get?track_name=Blinding%20Lights&artist_name=The%20Weeknd
```

**Response (200 OK):**
```json
{
  "id": 12345,
  "trackName": "Blinding Lights",
  "artistName": "The Weeknd",
  "albumName": "After Hours",
  "duration": 200.5,
  "plainLyrics": "I've been tryna call...",
  "syncedLyrics": "[00:10.00]I've been tryna call..."
}
```

---

### GET /api/search

Fuzzy search for lyrics.

**Parameters:**
| Name | Required | Description |
|------|----------|-------------|
| `track_name` | No | Song title |
| `artist_name` | No | Artist name |
| `q` | No | General query |

**Example:**
```
GET /api/search?track_name=Blinding%20Lights&artist_name=The%20Weeknd
GET /api/search?q=Blinding%20Lights%20Weeknd
```

**Response (200 OK):**
```json
[
  {
    "id": 12345,
    "trackName": "Blinding Lights",
    "artistName": "The Weeknd",
    "albumName": "After Hours",
    "duration": 200.5,
    "syncedLyrics": "...",
    "plainLyrics": "..."
  }
]
```

---

### GET /api/get/{id}

Fetch lyrics by LRCLIB ID.

**Example:**
```
GET /api/get/12345
```

**Response:** Same as `/api/get`

---

## LRC Format

Synced lyrics use LRC format with timestamps:

```
[mm:ss.xx]Lyrics line text
```

**Examples:**
```
[00:10.00]First line of the song
[00:15.50]Second line here
[01:02.34]Later in the song
```

**Parsing:** See [lyrics package](backend/lyrics.md) for implementation.

---

## See Also

- [Lyrics Package](backend/lyrics.md) - LRCLIB implementation
- [Auth Package](backend/auth.md) - Spotify OAuth
- [Glossary](GLOSSARY.md) - LRC format explanation

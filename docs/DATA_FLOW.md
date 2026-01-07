# Data Flow

How data moves through the SpotLy application.

---

## Authentication Flow

```mermaid
sequenceDiagram
    participant User
    participant SpotLy
    participant Browser
    participant Spotify
    
    User->>SpotLy: Click "Connect with Spotify"
    SpotLy->>SpotLy: Start callback server (:8080)
    SpotLy->>Browser: Open auth URL
    Browser->>Spotify: User logs in
    Spotify->>Browser: Redirect to /callback?code=xxx
    Browser->>SpotLy: GET /callback?code=xxx
    SpotLy->>Spotify: Exchange code for tokens
    Spotify->>SpotLy: Access + Refresh tokens
    SpotLy->>SpotLy: Save tokens to config
    SpotLy->>SpotLy: Create Spotify client
    SpotLy->>User: Show overlay UI
```

---

## Lyrics Fetching Pipeline

```mermaid
flowchart TB
    A[Track Change Detected] --> B{Cache by Track ID?}
    B -->|Hit| C[Return cached lyrics]
    B -->|Miss| D{Cache by Artist+Title?}
    D -->|Hit| E[Cache by Track ID too]
    E --> C
    D -->|Miss| F[Query LRCLIB API]
    F --> G{Exact match found?}
    G -->|Yes| H[Parse LRC timestamps]
    G -->|No| I{Search results?}
    I -->|Yes| J[Pick best match by score]
    J --> K[Fetch full lyrics by ID]
    K --> H
    I -->|No| L[Use Demo fallback]
    H --> M[Cache result]
    M --> C
    L --> C
```

---

## Real-Time Display Update

```mermaid
flowchart LR
    subgraph "Backend (5s interval)"
        A[Spotify Poll] --> B[Update TrackInfo]
        B --> C[Progress + Timestamp]
    end
    
    subgraph "Frontend (50ms interval)"
        D[GetDisplayInfo()] --> E[Calculate current progress]
        E --> F[Find matching lyrics line]
        F --> G[Update UI]
    end
    
    C --> D
```

### Progress Calculation

```
effective_progress = stored_progress 
                   + time_since_last_poll 
                   + sync_offset
```

This allows smooth karaoke animation without waiting for API polls.

---

## Configuration Persistence

```mermaid
flowchart TB
    A[App Start] --> B{config.json exists?}
    B -->|Yes| C[Load config]
    B -->|No| D[Create default config]
    D --> E[Save to disk]
    C --> F[App Running]
    E --> F
    
    F --> G[User changes setting]
    G --> H[Update config in memory]
    H --> I[Save to disk]
    I --> F
    
    F --> J[App Shutdown]
    J --> K[Final save to disk]
```

---

## Service Dependencies

```mermaid
graph TD
    Config --> Auth
    Config --> Overlay
    Cache --> Lyrics
    Auth --> Spotify
    Overlay --> Spotify
    Lyrics --> Spotify
    
    subgraph "Initialization Order"
        1[Config] --> 2[Cache]
        2 --> 3[Overlay]
        3 --> 4[Auth]
        4 --> 5[Lyrics]
        5 --> 6[Spotify]
    end
```

---

## Frontend ↔ Backend Communication

| Direction | Mechanism | Example |
|-----------|-----------|---------|
| Frontend → Backend | Wails bindings | `window.go.main.App.GetDisplayInfo()` |
| Backend → Frontend | Return values | `DisplayInfo` struct as JSON |
| Direct UI control | Wails runtime | `window.runtime.WindowSetPosition()` |

---

## See Also

- [Architecture](ARCHITECTURE.md) - System overview
- [Backend Overview](backend/README.md) - Package details
- [API Reference](API_REFERENCE.md) - External API details

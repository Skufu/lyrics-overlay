# SpotLy Overlay - Codebase Overview

## Project Overview

**SpotLy Overlay** is a desktop application that displays time-synced or plain text lyrics for currently playing Spotify tracks in an always-on-top, transparent overlay window. Built using Go and Wails v2, it provides a gaming-friendly lyrics display that automatically becomes click-through when games like VALORANT are active.

## Core Functionality

### 🎵 **Lyrics Display**
- **Real-time synchronization**: Shows lyrics synced with Spotify playback progress
- **Multiple providers**: LRCLIB (primary) and Genius (fallback) lyrics sources
- **Fallback system**: Demo lyrics when no sources are available
- **Smart caching**: LRU cache for lyrics by track ID and normalized artist/title

### 🎮 **Gaming Integration**
- **Click-through overlay**: Automatically becomes transparent to mouse events during games
- **VALORANT detection**: Specifically optimized for gaming compatibility
- **Always-on-top**: Stays visible above fullscreen applications

### 🔐 **Authentication**
- **Spotify OAuth2**: Authorization Code flow with localhost callback
- **Token persistence**: Automatic refresh and storage in user config
- **Secure handling**: Proper token lifecycle management

## Architecture

### **Technology Stack**
- **Backend**: Go 1.24.1 with Wails v2 framework
- **Frontend**: HTML/CSS/JavaScript (embedded in binary)
- **Desktop Integration**: Windows API for window management and click-through functionality
- **External APIs**: Spotify Web API, LRCLIB API, Genius API

### **Project Structure**

```
lyrics-overlay/
├── main.go                 # Wails application entry point & lifecycle management
├── wails.json             # Wails build configuration
├── go.mod/go.sum          # Go module dependencies
├── build.bat              # Windows build script
├── frontend/dist/         # Compiled frontend assets
├── internal/
│   ├── auth/              # Spotify OAuth2 implementation
│   ├── cache/             # LRU cache for lyrics
│   ├── config/            # Configuration persistence
│   ├── lyrics/            # Multi-provider lyrics fetching
│   ├── overlay/           # Window management & display logic
│   └── spotify/           # Spotify API client & polling
└── README.md              # Project documentation
```

## Key Components

### **1. Main Application (`main.go`)**
- **Wails lifecycle management**: Startup, shutdown, and window configuration
- **Service orchestration**: Coordinates all internal services
- **Windows integration**: Click-through monitoring and window management
- **Configuration**: Determines startup options (resize disable, etc.)

### **2. Authentication (`internal/auth/`)**
- **OAuth2 flow**: Handles Spotify authorization with localhost callback
- **Token management**: Secure storage and automatic refresh
- **Client creation**: Provides authenticated Spotify API client

### **3. Lyrics System (`internal/lyrics/`)**
- **Multi-provider architecture**: LRCLIB → Genius → Demo fallback chain
- **Smart caching**: Prevents redundant API calls
- **Text normalization**: Handles common metadata variations
- **LRC parsing**: Converts synced lyrics to timestamped format

### **4. Overlay Management (`internal/overlay/`)**
- **Window positioning**: Configurable position and sizing
- **Display state**: Current line, next line, playback status
- **Track information**: Artist, title, progress, and metadata
- **UI controls**: Visibility, opacity, font size, locking

### **5. Spotify Integration (`internal/spotify/`)**
- **Resilient polling**: Exponential backoff for API rate limiting
- **Real-time updates**: Track changes and playback state
- **Error handling**: Graceful degradation on API failures

### **6. Configuration (`internal/config/`)**
- **User preferences**: Overlay settings, API keys, window position
- **Persistent storage**: JSON configuration in `~/.spotly/`
- **Runtime updates**: Hot-reload configuration changes

## Data Flow

### **Application Lifecycle**
1. **Startup**: Initialize services → Load config → Check authentication
2. **Authentication**: OAuth flow → Token storage → Spotify client creation
3. **Polling**: Real-time track monitoring → Lyrics fetching → Cache updates
4. **Display**: Overlay positioning → Lyrics rendering → UI updates
5. **Gaming Mode**: Active window detection → Click-through toggle

### **Lyrics Retrieval Process**
1. **Cache Check**: Track ID → Normalized artist/title lookup
2. **Provider Chain**: LRCLIB (synced) → Genius (fallback) → Demo (info only)
3. **Text Processing**: HTML parsing → Noise filtering → Line formatting
4. **Display Sync**: Timestamp alignment → Progress tracking

## Configuration

### **User Settings (`~/.spotly/config.json`)**
```json
{
  "spotify_client_id": "your_spotify_client_id",
  "spotify_client_secret": "your_spotify_client_secret",
  "redirect_uri": "http://127.0.0.1:8080/callback",
  "genius_token": "optional_genius_api_token",
  "overlay": {
    "x": 100, "y": 100, "width": 600, "height": 120,
    "opacity": 0.9, "font_size": 16, "visible": true,
    "locked": false, "position": "bottom-left"
  }
}
```

## Dependencies

### **Core Libraries**
- **Wails v2**: Desktop application framework
- **Spotify Web API Go**: Official Spotify API client
- **golang.org/x/oauth2**: OAuth2 implementation
- **golang.org/x/net**: HTML parsing for Genius lyrics

### **Key Features**
- **LRU Cache**: Efficient memory usage with intelligent eviction
- **Exponential Backoff**: API rate limiting compliance
- **HTML Parsing**: Robust lyrics extraction from Genius pages
- **Windows API**: Native window management and click-through

## Development Status

### **✅ Implemented**
- Core lyrics overlay functionality
- Spotify OAuth2 integration
- Multi-provider lyrics fetching
- Gaming click-through support
- Configuration persistence
- LRU caching system

### **🔄 In Progress**
- Global hotkeys implementation
- Enhanced lyrics scraping
- Dependency resolution

### **📋 Planned**
- Windows packaging and testing
- Enhanced UI controls
- Cross-platform support (macOS/Linux)

## API Integrations

### **Spotify Web API**
- `GET /me/player/currently-playing`: Track and progress data
- `GET /me/player`: Enhanced device information
- **Rate Limiting**: 3-5 second polling intervals with exponential backoff

### **LRCLIB API**
- `GET /api/get`: Direct track lookup with synced/plain lyrics
- `GET /api/search`: Fallback search functionality
- **Priority**: Preferred source for synced lyrics

### **Genius API**
- Search endpoint: Song discovery and matching
- Web scraping: Lyrics extraction from song pages
- **Fallback**: Used when LRCLIB doesn't have the track

## Build & Deployment

### **Development**
```bash
go mod tidy                    # Install dependencies
wails dev                     # Development server
```

### **Production**
```bash
wails build                   # Cross-platform binary
./build.bat                   # Windows quick build
```

### **Output**
- **Binary**: `spotly.exe` (Windows)
- **Assets**: Embedded frontend in binary
- **Config**: `~/.spotly/config.json` (auto-created)

## Error Handling

### **Resilience Features**
- **Provider fallback**: Automatic failover between lyrics sources
- **Cache degradation**: Graceful handling of cache misses
- **API timeouts**: Configurable request timeouts (30s default)
- **Rate limiting**: Exponential backoff prevents API bans

### **Common Issues**
- **Authentication failures**: Config validation and clear error messages
- **No lyrics found**: Multiple providers with intelligent fallbacks
- **Gaming interference**: Click-through detection and auto-enabling
- **Performance**: Efficient polling and caching strategies

## Performance Characteristics

### **Memory Usage**
- **LRU Cache**: 100 entry limit with intelligent eviction
- **Provider instances**: Reused HTTP clients and connections
- **Asset embedding**: Frontend assets compiled into binary

### **Network Efficiency**
- **Connection reuse**: Persistent HTTP clients
- **Smart caching**: Prevents redundant API calls
- **Batch processing**: Efficient lyrics parsing and formatting

### **CPU Usage**
- **Polling intervals**: 3-5 second Spotify API checks
- **Background monitoring**: 500ms active window detection
- **On-demand processing**: Lyrics fetched only when tracks change

## Security Considerations

### **Authentication**
- **OAuth2 PKCE**: Industry-standard authorization flow
- **Token storage**: Local filesystem with appropriate permissions
- **Scope limitation**: Minimal required Spotify permissions

### **API Keys**
- **Environment handling**: Secure storage recommendations
- **Optional dependencies**: Genius token not required for basic functionality

## Testing Strategy

### **Integration Points**
- **Spotify API**: Real API calls with proper mocking
- **Lyrics providers**: Multiple source validation
- **Windows integration**: Native API testing
- **Configuration**: File I/O and persistence testing

### **Edge Cases**
- **Network failures**: Offline and timeout handling
- **Rate limiting**: API quota management
- **Gaming mode**: Fullscreen application detection
- **Invalid data**: Malformed lyrics and metadata handling

This codebase represents a well-architected, gaming-focused lyrics overlay application with robust error handling, intelligent caching, and seamless desktop integration.

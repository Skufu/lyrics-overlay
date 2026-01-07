# SpotLy Documentation

> Comprehensive documentation for the SpotLy lyrics overlay application.

## Quick Start

- **New to SpotLy?** → [Development Guide](DEVELOPMENT.md)
- **Understanding the code?** → [Architecture Overview](ARCHITECTURE.md)
- **Having issues?** → [Troubleshooting](TROUBLESHOOTING.md)

---

## Documentation Index

### Core Documentation

| Document | Description |
|----------|-------------|
| [Architecture](ARCHITECTURE.md) | System overview, components, and technology stack |
| [Data Flow](DATA_FLOW.md) | How data moves through authentication, lyrics, and display |
| [Configuration](CONFIGURATION.md) | Complete config.json reference |
| [Development](DEVELOPMENT.md) | Setup, build, and contribution guide |

### Reference

| Document | Description |
|----------|-------------|
| [API Reference](API_REFERENCE.md) | Spotify and LRCLIB API documentation |
| [Platform Support](PLATFORM_SUPPORT.md) | Windows features, cross-platform status |
| [Testing](TESTING.md) | Test coverage and commands |
| [Troubleshooting](TROUBLESHOOTING.md) | Common issues and solutions |
| [Glossary](GLOSSARY.md) | Terms and concepts |

### Component Documentation

| Section | Description |
|---------|-------------|
| [Backend](backend/README.md) | Go packages: auth, cache, config, lyrics, overlay, spotify |
| [Frontend](frontend/README.md) | HTML/JS/CSS overlay UI |
| [Wails Integration](wails/README.md) | Desktop framework and bindings |

---

## Project Overview

**SpotLy** is a transparent, always-on-top lyrics overlay for Spotify, built for gaming.

```
┌─────────────────────────────────────────────────────────────┐
│                      SpotLy Overlay                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │            ♪ Current lyrics line here ♪              │   │
│  │              Next line preview...                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Key Features

- **Real-time synced lyrics** with karaoke-style highlighting
- **Transparent overlay** that stays on top of games
- **Auto click-through** when gaming (Windows)
- **LRCLIB integration** for timestamped lyrics

---

## See Also

- [GitHub Repository](https://github.com/Skufu/lyrics-overlay)
- [README.md](../README.md) - Project readme with installation instructions

# SpotLy

<p>
  <a href="https://github.com/Skufu/lyrics-overlay/actions/workflows/ci.yml"><img src="https://github.com/Skufu/lyrics-overlay/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/Skufu/lyrics-overlay/releases"><img src="https://img.shields.io/github/release/Skufu/lyrics-overlay.svg" alt="Latest Release"></a>
  <a href="https://github.com/Skufu/lyrics-overlay/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/Skufu/lyrics-overlay"><img src="https://goreportcard.com/badge/github.com/Skufu/lyrics-overlay" alt="Go Report Card"></a>
</p>

a transparent, always-on-top lyrics overlay for spotify. i built this because i wanted lyrics on screen while gaming without alt-tabbing or dealing with bloated apps. no album art, no branding, no clutter — just the words.

## demo


<video src="spotly-demo.webm" width="100%" controls></video>

## what it does

- syncs lyrics to your spotify playback in real time
- karaoke-style highlighting so you always know where you are in the song
- stays on top of fullscreen games (borderless windowed)
- fully transparent background, blends right into whatever you're doing
- drag it wherever you want, lock it in place, tweak opacity & font size

basically, it just sits there and shows you lyrics. that's it. that's the app.


## getting started

### download

grab `spotly.exe` from [releases](https://github.com/Skufu/lyrics-overlay/releases) and run it. done.

### build from source

if you want to build it yourself:

```bash
# you'll need the wails cli
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# clone and build
git clone https://github.com/Skufu/lyrics-overlay.git
cd lyrics-overlay
wails build
```

needs Go 1.22+ and a [Spotify Developer Account](https://developer.spotify.com/dashboard).


## setup

### 1. create a spotify app

head to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard) and make a new app. under **Redirect URIs**, add:

```
http://127.0.0.1:58432/callback
```

copy your **Client ID** and **Client Secret**.

### 2. launch spotly

run the exe. first time around, it'll ask for your spotify credentials. paste in your Client ID and Secret, hit **Connect with Spotify**.

### 3. authenticate

your browser opens, you authorize with spotify, and you're good. lyrics start showing up for whatever you're playing.


## controls

hover over the overlay to see these:

| control | what it does |
|---------|-------------|
| drag | move the window around |
| lock icon | lock it so you don't accidentally nudge it |
| settings | font size, opacity, sync offset |
| refresh | re-fetch the current track |

### sync offset

if the lyrics are slightly off from the music, tweak the timing slider in settings:

- **positive** = lyrics show earlier
- **negative** = lyrics show later

350ms default works for most setups.


## config

everything lives in `~/.spotly/config.json` (on windows: `C:\Users\<you>\.spotly\config.json`):

```json
{
  "spotify_client_id": "your_client_id",
  "spotify_client_secret": "your_client_secret",
  "redirect_uri": "http://127.0.0.1:58432/callback",
  "port": 58432,
  "overlay": {
    "x": 100,
    "y": 100,
    "width": 600,
    "height": 120,
    "opacity": 0.9,
    "font_size": 16,
    "visible": true,
    "locked": false,
    "position": "bottom-left",
    "sync_offset": 350
  }
}
```


## how it's built

```
spotly/
├── main.go                 # wails app entry
├── internal/
│   ├── auth/               # spotify oauth2
│   ├── cache/              # lru lyrics cache
│   ├── config/             # config persistence
│   ├── lyrics/             # lrclib provider
│   ├── overlay/            # display state
│   └── spotify/            # api client & polling
└── frontend/dist/          # overlay ui
```

uses the [Spotify Web API](https://developer.spotify.com/documentation/web-api/) for playback tracking and [LRCLIB](https://lrclib.net/) for synced lyrics. built with [Wails](https://wails.io/) so it's a proper native window, not an electron blob.


## troubleshooting

**oauth callback fails**
- redirect uri has to match exactly: `http://127.0.0.1:58432/callback`
- make sure port 58432 isn't being used by something else
- try disabling firewall if it still won't work

**no lyrics showing up**
- lrclib covers most popular songs but not everything
- some tracks just don't have synced lyrics available

**can't see the overlay in fullscreen**
- use borderless windowed mode in your game
- some anti-cheat systems (vanguard, etc.) block overlays

**build issues**
```bash
wails doctor        # check your deps
go mod tidy         # fix module stuff
```


## roadmap

stuff i want to add eventually:

- [ ] global hotkeys
- [ ] macos & linux support
- [ ] multi-monitor positioning
- [ ] custom themes


## contributing

if you want to contribute, go for it. PRs welcome.

things i'd especially appreciate help with:
- cross-platform support (mac/linux)
- global hotkey implementation
- performance stuff


## shoutouts

- [Wails](https://wails.io/) — go desktop framework that actually works
- [LRCLIB](https://lrclib.net/) — free synced lyrics api, absolute lifesaver
- [Spotify Web API](https://developer.spotify.com/documentation/web-api/) — does what it says


## license

[MIT](LICENSE) — do whatever you want with it.

---

made for my gaming setup. hope it's useful for yours too ✌️

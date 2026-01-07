# Troubleshooting

Common issues and solutions for SpotLy.

---

## OAuth Issues

### "Auth service not initialized"

**Cause:** Spotify credentials are missing or invalid.

**Solution:**
1. Run the Setup Wizard from the auth screen
2. Or manually edit `~/.spotly/config.json`:
   ```json
   {
     "spotify_client_id": "your_id_here",
     "spotify_client_secret": "your_secret_here"
   }
   ```

---

### OAuth callback fails / "Invalid redirect URI"

**Cause:** Redirect URI mismatch between app and Spotify dashboard.

**Solution:**
1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Edit your app settings
3. Ensure redirect URI is exactly: `http://127.0.0.1:8080/callback`
4. Save changes

---

### Port 8080 already in use

**Cause:** Another application is using port 8080.

**Solution:**
1. Find and stop the conflicting application
2. Or change port in config:
   ```json
   {
     "port": 9090,
     "redirect_uri": "http://127.0.0.1:9090/callback"
   }
   ```
3. **Important:** Update redirect URI in Spotify dashboard too!

---

## Lyrics Issues

### No lyrics found

**Possible causes:**
- Song too new or obscure
- LRCLIB doesn't have the song
- Artist/title mismatch

**Solutions:**
1. Click Refresh button
2. Check if lyrics are available on [lrclib.net](https://lrclib.net)
3. Some songs genuinely don't have lyrics available

---

### Lyrics out of sync

**Cause:** Default timing doesn't match your setup.

**Solution:**
1. Open Settings panel
2. Adjust Sync Offset slider
3. Positive = lyrics earlier, Negative = lyrics later
4. Default is +350ms

---

### Demo/Info lyrics showing instead of real lyrics

**Cause:** Cache returned fallback result.

**Solution:**
1. Click Refresh button
2. Wait a few seconds and check again
3. Demo lyrics are not cached, so a refresh should re-fetch

---

## Overlay Issues

### Overlay not visible in fullscreen game

**Common causes:**
1. Game using exclusive fullscreen mode
2. Anti-cheat blocking overlays

**Solutions:**
- Use **Borderless Windowed** mode instead of Fullscreen
- Some anti-cheat systems (Vanguard, EasyAntiCheat) may block overlays

---

### Can't click through overlay in game

**Cause:** Click-through mode not activating.

**Solution:**
1. Ensure game is in the supported list:
   - VALORANT
   - League of Legends
   - CS2
   - Dota 2
   - Overwatch
   - Apex Legends
2. Click-through activates when game window is focused
3. When in desktop, overlay is clickable (expected)

---

### Overlay stuck / can't move

**Cause:** Overlay is locked.

**Solution:**
1. Click on the lyrics to show the gear icon
2. Open Settings
3. Click "Lock" button to toggle off

---

## Build Issues

### "wails: command not found"

**Solution:**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

---

### Build fails with WebView2 error

**Cause:** Missing WebView2 runtime.

**Solution:**
- Usually pre-installed on Windows 10/11
- Download from [Microsoft](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)

---

### wails doctor shows issues

**Solution:**
```bash
wails doctor
go mod tidy
```

---

## See Also

- [Configuration](CONFIGURATION.md) - Config file reference
- [Development](DEVELOPMENT.md) - Build instructions
- [Platform Support](PLATFORM_SUPPORT.md) - Windows features

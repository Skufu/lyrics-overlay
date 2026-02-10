package overlay

import (
	"log"
	"testing"
	"time"

	"lyrics-overlay/internal/config"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	cfgSvc, err := config.New()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}
	svc, err := New(cfgSvc)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	return svc
}

// --- No Track / Loading States ---

func TestGetDisplayInfo_NoTrack(t *testing.T) {
	svc := newTestService(t)

	info := svc.GetDisplayInfo()
	if info.IsPlaying {
		t.Error("Expected IsPlaying false when no track")
	}
	if info.CurrentLine != "No track playing" {
		t.Errorf("Expected 'No track playing', got '%s'", info.CurrentLine)
	}
}

func TestGetDisplayInfo_TrackNoLyrics(t *testing.T) {
	svc := newTestService(t)

	svc.SetCurrentTrack(&TrackInfo{
		ID:        "123",
		Name:      "Test Song",
		IsPlaying: true,
		UpdatedAt: time.Now(),
	})
	svc.SetCurrentLyrics(nil)

	info := svc.GetDisplayInfo()
	if !info.IsPlaying {
		t.Error("Expected IsPlaying true")
	}
	if !info.IsLoading {
		t.Error("Expected IsLoading true when lyrics are nil")
	}
	if info.CurrentLine != "Fetching lyrics..." {
		t.Errorf("Expected 'Fetching lyrics...', got '%s'", info.CurrentLine)
	}
}

func TestGetDisplayInfo_EmptyLyrics(t *testing.T) {
	svc := newTestService(t)

	svc.SetCurrentTrack(&TrackInfo{
		ID:        "123",
		Name:      "Instrumental",
		IsPlaying: true,
		UpdatedAt: time.Now(),
	})
	svc.SetCurrentLyrics(&LyricsData{
		Lines:    []LyricsLine{},
		IsSynced: false,
	})

	info := svc.GetDisplayInfo()
	if info.CurrentLine != "No lyrics available" {
		t.Errorf("Expected 'No lyrics available', got '%s'", info.CurrentLine)
	}
}

// --- Non-synced Lyrics ---

func TestGetDisplayInfo_NonSyncedLyrics(t *testing.T) {
	svc := newTestService(t)

	svc.SetCurrentTrack(&TrackInfo{
		ID:        "456",
		Name:      "Plain Lyrics Song",
		IsPlaying: true,
		UpdatedAt: time.Now(),
	})
	svc.SetCurrentLyrics(&LyricsData{
		Lines: []LyricsLine{
			{Text: "First line of lyrics"},
			{Text: "Second line of lyrics"},
			{Text: "Third line of lyrics"},
		},
		IsSynced: false,
	})

	info := svc.GetDisplayInfo()
	if info.CurrentLine != "First line of lyrics" {
		t.Errorf("Expected first line, got '%s'", info.CurrentLine)
	}
	if info.NextLine != "Second line of lyrics" {
		t.Errorf("Expected second line as next, got '%s'", info.NextLine)
	}
}

// --- Synced Lyrics: Line Selection ---

func TestGetDisplayInfo_SyncedLyrics_FirstLine(t *testing.T) {
	svc := newTestService(t)

	// Set track at beginning (progress 0ms, just started)
	svc.SetCurrentTrack(&TrackInfo{
		ID:        "sync1",
		Name:      "Synced Song",
		Progress:  500, // 500ms into the song
		Duration:  180000,
		IsPlaying: false, // not playing so elapsed=0
		UpdatedAt: time.Now(),
	})

	svc.SetCurrentLyrics(&LyricsData{
		Lines: []LyricsLine{
			{Text: "Intro line", Timestamp: 0},
			{Text: "Second verse", Timestamp: 5000},
			{Text: "Third verse", Timestamp: 10000},
		},
		IsSynced: true,
	})

	info := svc.GetDisplayInfo()
	// At 500ms + sync offset (~350ms) = ~850ms, should be on first line (ts=0)
	if info.CurrentLine != "Intro line" {
		t.Errorf("Expected 'Intro line' at early progress, got '%s'", info.CurrentLine)
	}
	if info.NextLine != "Second verse" {
		t.Errorf("Expected 'Second verse' as next, got '%s'", info.NextLine)
	}
}

func TestGetDisplayInfo_SyncedLyrics_MidSong(t *testing.T) {
	svc := newTestService(t)

	// Track at 7000ms — should show second line (ts=5000)
	svc.SetCurrentTrack(&TrackInfo{
		ID:        "sync2",
		Name:      "Mid Song",
		Progress:  7000,
		Duration:  180000,
		IsPlaying: false, // not playing so elapsed=0
		UpdatedAt: time.Now(),
	})

	svc.SetCurrentLyrics(&LyricsData{
		Lines: []LyricsLine{
			{Text: "Intro", Timestamp: 0},
			{Text: "Verse one", Timestamp: 5000},
			{Text: "Chorus", Timestamp: 10000},
			{Text: "Verse two", Timestamp: 15000},
		},
		IsSynced: true,
	})

	info := svc.GetDisplayInfo()
	// At 7000ms + 350ms offset = 7350ms, between line 2 (5000) and line 3 (10000)
	if info.CurrentLine != "Verse one" {
		t.Errorf("Expected 'Verse one' at 7000ms, got '%s'", info.CurrentLine)
	}
	if info.NextLine != "Chorus" {
		t.Errorf("Expected 'Chorus' as next line, got '%s'", info.NextLine)
	}
}

func TestGetDisplayInfo_SyncedLyrics_LastLine(t *testing.T) {
	svc := newTestService(t)

	// Track near end — should show last line
	svc.SetCurrentTrack(&TrackInfo{
		ID:        "sync3",
		Name:      "Last Line Song",
		Progress:  20000,
		Duration:  25000,
		IsPlaying: false,
		UpdatedAt: time.Now(),
	})

	svc.SetCurrentLyrics(&LyricsData{
		Lines: []LyricsLine{
			{Text: "Intro", Timestamp: 0},
			{Text: "Middle", Timestamp: 10000},
			{Text: "Outro", Timestamp: 18000},
		},
		IsSynced: true,
	})

	info := svc.GetDisplayInfo()
	// At 20000ms + 350ms = 20350ms, past last line timestamp (18000ms)
	if info.CurrentLine != "Outro" {
		t.Errorf("Expected 'Outro' at end of song, got '%s'", info.CurrentLine)
	}
	// No more lines after — next should be empty
	if info.NextLine != "" {
		t.Errorf("Expected empty next line at end, got '%s'", info.NextLine)
	}
}

// --- Synced Lyrics: Empty Line Handling ---

func TestGetDisplayInfo_SyncedLyrics_SkipsEmptyLines(t *testing.T) {
	svc := newTestService(t)

	// Progress is at an empty/instrumental break line
	svc.SetCurrentTrack(&TrackInfo{
		ID:        "sync4",
		Name:      "Song With Breaks",
		Progress:  12000,
		Duration:  60000,
		IsPlaying: false,
		UpdatedAt: time.Now(),
	})

	svc.SetCurrentLyrics(&LyricsData{
		Lines: []LyricsLine{
			{Text: "Verse one", Timestamp: 0},
			{Text: "", Timestamp: 10000}, // Instrumental break
			{Text: "Verse two", Timestamp: 20000},
		},
		IsSynced: true,
	})

	info := svc.GetDisplayInfo()
	// At 12000ms + 350ms = 12350ms, current timestamp is the empty line at 10000ms
	// Empty lines should be skipped, showing next non-empty line
	if info.CurrentLine != "Verse two" {
		t.Errorf("Expected empty line to skip to 'Verse two', got '%s'", info.CurrentLine)
	}
}

// --- Karaoke Progress Math ---

func TestGetDisplayInfo_KaraokeProgress(t *testing.T) {
	svc := newTestService(t)

	// Track halfway through a known line
	svc.SetCurrentTrack(&TrackInfo{
		ID:        "karaoke1",
		Name:      "Karaoke Song",
		Progress:  7000, // 7s into song
		Duration:  60000,
		IsPlaying: false,
		UpdatedAt: time.Now(),
	})

	svc.SetCurrentLyrics(&LyricsData{
		Lines: []LyricsLine{
			{Text: "First line", Timestamp: 5000},
			{Text: "Second line", Timestamp: 10000},
		},
		IsSynced: true,
	})

	info := svc.GetDisplayInfo()
	// Effective progress: 7000 + 350 = 7350ms
	// Current line starts at 5000ms, next at 10000ms
	// LineDuration should be 10000 - 5000 = 5000ms
	// LineProgress should be 7350 - 5000 = 2350ms

	if info.LineDuration != 5000 {
		t.Errorf("Expected LineDuration 5000, got %d", info.LineDuration)
	}
	if info.LineProgress < 2000 || info.LineProgress > 3000 {
		t.Errorf("Expected LineProgress ~2350, got %d", info.LineProgress)
	}
	if info.LineStartTime != 5000 {
		t.Errorf("Expected LineStartTime 5000, got %d", info.LineStartTime)
	}
}

func TestGetDisplayInfo_KaraokeProgress_Clamped(t *testing.T) {
	svc := newTestService(t)

	// Progress before line start — lineProgress should be 0
	svc.SetCurrentTrack(&TrackInfo{
		ID:        "karaoke2",
		Name:      "Clamp Test",
		Progress:  100,
		Duration:  60000,
		IsPlaying: false,
		UpdatedAt: time.Now(),
	})

	svc.SetCurrentLyrics(&LyricsData{
		Lines: []LyricsLine{
			{Text: "First line", Timestamp: 0},
			{Text: "Second line", Timestamp: 5000},
		},
		IsSynced: true,
	})

	info := svc.GetDisplayInfo()
	// LineProgress should be >= 0 (clamped)
	if info.LineProgress < 0 {
		t.Errorf("Expected LineProgress >= 0, got %d", info.LineProgress)
	}
	// And should not exceed LineDuration
	if info.LineDuration > 0 && info.LineProgress > info.LineDuration {
		t.Errorf("Expected LineProgress <= LineDuration, got %d > %d", info.LineProgress, info.LineDuration)
	}
}

// --- Visibility ---

func TestService_Visibility(t *testing.T) {
	svc := newTestService(t)

	// Force visible
	svc.SetVisibility(true)
	if !svc.IsVisible() {
		t.Error("Expected visible after SetVisibility(true)")
	}

	// Toggle
	visible := svc.ToggleVisibility()
	if visible {
		t.Error("Expected hidden after toggle")
	}
	if svc.IsVisible() {
		t.Error("Expected IsVisible() false after toggle")
	}

	// Set back
	svc.SetVisibility(true)
	if !svc.IsVisible() {
		t.Error("Expected visible after SetVisibility(true)")
	}
}

// --- Track and Lyrics Management ---

func TestService_SetAndGetTrack(t *testing.T) {
	svc := newTestService(t)

	if svc.GetCurrentTrack() != nil {
		t.Error("Expected nil track initially")
	}

	track := &TrackInfo{
		ID:      "track1",
		Name:    "Test",
		Artists: []string{"Artist"},
	}
	svc.SetCurrentTrack(track)

	got := svc.GetCurrentTrack()
	if got == nil {
		t.Fatal("Expected non-nil track")
	}
	if got.ID != "track1" {
		t.Errorf("Expected ID 'track1', got '%s'", got.ID)
	}

	// Clear track
	svc.SetCurrentTrack(nil)
	if svc.GetCurrentTrack() != nil {
		t.Error("Expected nil after clearing track")
	}
}

func TestService_SetAndGetLyrics(t *testing.T) {
	svc := newTestService(t)

	if svc.GetCurrentLyrics() != nil {
		t.Error("Expected nil lyrics initially")
	}

	lyrics := &LyricsData{
		Lines:    []LyricsLine{{Text: "Hello"}},
		IsSynced: false,
	}
	svc.SetCurrentLyrics(lyrics)

	got := svc.GetCurrentLyrics()
	if got == nil {
		t.Fatal("Expected non-nil lyrics")
	}
	if len(got.Lines) != 1 || got.Lines[0].Text != "Hello" {
		t.Errorf("Unexpected lyrics: %+v", got)
	}
}

// --- Config ---

func TestService_UpdateOverlayConfig(t *testing.T) {
	svc := newTestService(t)

	cfg := svc.GetOverlayConfig()
	cfg.FontSize = 24
	cfg.KaraokeEnabled = true

	err := svc.UpdateOverlayConfig(cfg)
	if err != nil {
		t.Fatalf("UpdateOverlayConfig failed: %v", err)
	}

	updated := svc.GetOverlayConfig()
	if updated.FontSize != 24 {
		t.Errorf("Expected FontSize 24, got %d", updated.FontSize)
	}
}

func init() {
	log.SetFlags(0)
}

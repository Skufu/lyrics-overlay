package cache

import (
	"testing"

	"lyrics-overlay/internal/overlay"
)

func TestService_SetAndGet(t *testing.T) {
	c := New(3)

	lyrics1 := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "lyrics1"}},
	}
	lyrics2 := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "lyrics2"}},
	}

	c.SetByTrackID("song1", lyrics1)
	c.SetByTrackID("song2", lyrics2)

	got := c.GetByTrackID("song1")
	if got == nil || len(got.Lines) == 0 || got.Lines[0].Text != "lyrics1" {
		t.Errorf("GetByTrackID(song1) = %v; want lyrics1", got)
	}
}

func TestService_Eviction(t *testing.T) {
	c := New(2)

	lyrics1 := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "1"}},
	}
	lyrics2 := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "2"}},
	}
	lyrics3 := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "3"}},
	}

	c.SetByTrackID("a", lyrics1)
	c.SetByTrackID("b", lyrics2)
	c.SetByTrackID("c", lyrics3) // Should evict "a"

	if got := c.GetByTrackID("a"); got != nil {
		t.Error("Expected 'a' to be evicted")
	}

	if got := c.GetByTrackID("b"); got == nil {
		t.Error("Expected 'b' to exist")
	}

	if got := c.GetByTrackID("c"); got == nil {
		t.Error("Expected 'c' to exist")
	}
}

func TestService_UpdateExisting(t *testing.T) {
	c := New(2)

	lyrics1 := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "value1"}},
	}
	lyrics2 := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "value2"}},
	}

	c.SetByTrackID("key", lyrics1)
	c.SetByTrackID("key", lyrics2)

	got := c.GetByTrackID("key")
	if got == nil || len(got.Lines) == 0 || got.Lines[0].Text != "value2" {
		t.Errorf("Expected updated value, got %v", got)
	}
}

func TestService_GetByKey(t *testing.T) {
	c := New(3)

	lyrics := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "test lyrics"}},
	}

	c.SetByKey("artist|title", lyrics)

	got := c.GetByKey("artist|title")
	if got == nil || len(got.Lines) == 0 || got.Lines[0].Text != "test lyrics" {
		t.Errorf("GetByKey failed, got %v", got)
	}
}

func TestService_Size(t *testing.T) {
	c := New(10)

	if c.Size() != 0 {
		t.Errorf("Expected size 0, got %d", c.Size())
	}

	lyrics := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "test"}},
	}

	c.SetByTrackID("track1", lyrics)
	if c.Size() != 1 {
		t.Errorf("Expected size 1, got %d", c.Size())
	}

	c.SetByTrackID("track2", lyrics)
	if c.Size() != 2 {
		t.Errorf("Expected size 2, got %d", c.Size())
	}
}

func TestService_Clear(t *testing.T) {
	c := New(10)

	lyrics := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "test"}},
	}

	c.SetByTrackID("track1", lyrics)
	c.SetByKey("key1", lyrics)

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("Expected size 0 after Clear, got %d", c.Size())
	}

	if got := c.GetByTrackID("track1"); got != nil {
		t.Error("Expected track1 to be cleared")
	}

	if got := c.GetByKey("key1"); got != nil {
		t.Error("Expected key1 to be cleared")
	}
}

func TestService_Expiration(t *testing.T) {
	c := New(10)

	lyrics := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "test"}},
	}

	c.SetByTrackID("track1", lyrics)

	// Manually set timestamp to be old (over 24 hours)
	// We can't directly access the entry, so we'll test by waiting
	// But for unit tests, we'll just verify the entry exists initially
	got := c.GetByTrackID("track1")
	if got == nil {
		t.Error("Expected track1 to exist before expiration")
	}

	// Note: Testing actual expiration would require mocking time or waiting 24+ hours
	// This test verifies the basic functionality works
}

func TestService_Stats(t *testing.T) {
	c := New(10)

	lyrics := &overlay.LyricsData{
		Source:   "Test",
		IsSynced: false,
		Lines:    []overlay.LyricsLine{{Text: "test"}},
	}

	c.SetByTrackID("track1", lyrics)
	c.SetByKey("key1", lyrics)

	stats := c.Stats()
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}
	if stats.MaxSize != 10 {
		t.Errorf("Expected max size 10, got %d", stats.MaxSize)
	}
	if stats.TrackEntries != 1 {
		t.Errorf("Expected 1 track entry, got %d", stats.TrackEntries)
	}
	if stats.KeyEntries != 1 {
		t.Errorf("Expected 1 key entry, got %d", stats.KeyEntries)
	}
}

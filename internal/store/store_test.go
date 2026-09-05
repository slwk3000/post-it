package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/slwk3000/post-it/internal/config"
	"github.com/slwk3000/post-it/internal/model"
)

func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "postit-store-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		DataDir:      tempDir,
		NotesFile:    filepath.Join(tempDir, "notes.json"),
		SettingsFile: filepath.Join(tempDir, "settings.json"),
	}

	store := New(cfg)
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return store, cleanup
}

func TestStoreNotesPersistence(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Initial load should be empty
	notes, err := s.LoadNotes()
	if err != nil {
		t.Fatalf("LoadNotes failed on empty: %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes, got %d", len(notes))
	}

	// Create and save notes
	note1 := &model.Note{
		ID:           "test-1",
		Content:      "First post-it note",
		X:            150,
		Y:            200,
		Width:        280,
		Height:       260,
		PaperType:    model.PaperPolen,
		PaperPattern: model.PatternDotted,
		PenColor:     model.PenBlue,
		Alignment:    model.AlignCorner,
		Color:        "#fcf5e5",
		Saturation:   80,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.SaveNotes([]*model.Note{note1}); err != nil {
		t.Fatalf("SaveNotes failed: %v", err)
	}

	loaded, err := s.LoadNotes()
	if err != nil {
		t.Fatalf("LoadNotes failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 note, got %d", len(loaded))
	}
	if loaded[0].Content != "First post-it note" {
		t.Errorf("content mismatch: got %s", loaded[0].Content)
	}
	if loaded[0].PaperType != model.PaperPolen {
		t.Errorf("paper type mismatch: got %s", loaded[0].PaperType)
	}
}

func TestStoreSettingsPersistence(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	settings, err := s.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed on empty: %v", err)
	}
	if settings.DefaultPaperType != model.PaperPolen {
		t.Errorf("expected default polen, got %s", settings.DefaultPaperType)
	}

	settings.DefaultPaperType = model.PaperKraft
	settings.DefaultPenColor = model.PenBlack
	if err := s.SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	reloaded, err := s.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}
	if reloaded.DefaultPaperType != model.PaperKraft {
		t.Errorf("expected kraft, got %s", reloaded.DefaultPaperType)
	}
	if reloaded.DefaultPenColor != model.PenBlack {
		t.Errorf("expected black, got %s", reloaded.DefaultPenColor)
	}
}

package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/slwk3000/post-it/internal/config"
	"github.com/slwk3000/post-it/internal/model"
)

type Store struct {
	cfg *config.Config
	mu  sync.RWMutex
}

func New(cfg *config.Config) *Store {
	return &Store{cfg: cfg}
}

func (s *Store) NotesFileExists() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, err := os.Stat(s.cfg.NotesFile)
	return err == nil
}

func (s *Store) LoadNotes() ([]*model.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.cfg.NotesFile)
	if os.IsNotExist(err) {
		return []*model.Note{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read notes file: %w", err)
	}

	var notes []*model.Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, fmt.Errorf("unmarshal notes: %w", err)
	}

	return notes, nil
}

func (s *Store) SaveNotes(notes []*model.Note) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return atomicWriteJSON(s.cfg.NotesFile, notes)
}

func (s *Store) LoadSettings() (*model.Settings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.cfg.SettingsFile)
	if os.IsNotExist(err) {
		return model.DefaultSettings(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read settings file: %w", err)
	}

	settings := model.DefaultSettings()
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, fmt.Errorf("unmarshal settings: %w", err)
	}

	return settings, nil
}

func (s *Store) SaveSettings(settings *model.Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return atomicWriteJSON(s.cfg.SettingsFile, settings)
}

func atomicWriteJSON(destPath string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	dir := filepath.Dir(destPath)
	tmpFile, err := os.CreateTemp(dir, "postit-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, destPath); err != nil {
		return fmt.Errorf("atomic rename: %w", err)
	}

	return nil
}

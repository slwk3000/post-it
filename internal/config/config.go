package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DataDir      string
	NotesFile    string
	SettingsFile string
}

func Default() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(home, ".config", "post-it")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	return &Config{
		DataDir:      dataDir,
		NotesFile:    filepath.Join(dataDir, "notes.json"),
		SettingsFile: filepath.Join(dataDir, "settings.json"),
	}, nil
}

package app

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DataDir   string
	DBPath    string
	IndexPath string
}

func NewConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	dataDir := filepath.Join(homeDir, ".bookmark-manager")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Config{
		DataDir:   dataDir,
		DBPath:    filepath.Join(dataDir, "bookmarks.db"),
		IndexPath: filepath.Join(dataDir, "search_index"),
	}, nil
}

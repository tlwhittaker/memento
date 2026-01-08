package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	APIBaseURL string
	AuthToken  string
}

func Load() (*Config, error) {
	loaded := false

	if homeDir, err := os.UserHomeDir(); err == nil {
		configEnvPath := filepath.Join(homeDir, ".config", "memento", ".env")
		if _, err := os.Stat(configEnvPath); err == nil {
			if err := godotenv.Load(configEnvPath); err == nil {
				loaded = true
			}
		}
	}

	if !loaded {
		_ = godotenv.Load()
	}

	apiURL := os.Getenv("MEMOS_API_URL")
	if apiURL == "" {
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".config", "memento", ".env")
		return nil, errors.New("MEMOS_API_URL is required. Create " + configPath + " with:\n" +
			"  MEMOS_API_URL=https://your-memos-instance.com/api/v1\n" +
			"  MEMOS_AUTH_TOKEN=your_access_token")
	}

	apiURL = strings.TrimSuffix(apiURL, "/")

	authToken := os.Getenv("MEMOS_AUTH_TOKEN")
	if authToken == "" {
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".config", "memento", ".env")
		return nil, errors.New("MEMOS_AUTH_TOKEN is required. Create " + configPath + " with:\n" +
			"  MEMOS_API_URL=https://your-memos-instance.com/api/v1\n" +
			"  MEMOS_AUTH_TOKEN=your_access_token")
	}

	return &Config{
		APIBaseURL: apiURL,
		AuthToken:  authToken,
	}, nil
}

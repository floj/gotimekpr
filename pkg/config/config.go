package config

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	DBURL            string
	TrackingInterval time.Duration
	NotifyBefore     time.Duration
	NoLogout         bool
}

func getXdgStateDir() (string, error) {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir != "" {
		return stateDir, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "state"), nil
}

func LoadConfig() (Config, error) {
	stateDir, err := getXdgStateDir()
	if err != nil {
		return Config{}, err
	}

	dbDir := filepath.Join(stateDir, "gotimekpr")
	if err := os.MkdirAll(dbDir, 0750); err != nil {
		return Config{}, err
	}

	dbURL := "file:" + filepath.Join(dbDir, "gotimekpr.db") + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	return Config{
		DBURL:            dbURL,
		TrackingInterval: 3 * time.Second,
		NotifyBefore:     1 * time.Minute,
	}, nil
}

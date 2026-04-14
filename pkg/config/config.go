package config

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	DBURL        string        `yaml:"db_url"`
	Interval     time.Duration `yaml:"interval"`
	NotifyBefore time.Duration `yaml:"notify_before"`
}

func LoadConfig() (Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}

	dbDir := filepath.Join(configDir, "gotimekpr")
	os.MkdirAll(dbDir, 0750)
	dbURL := "file:" + filepath.Join(dbDir, "gotimekpr.db") + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	return Config{
		DBURL:        dbURL,
		Interval:     5 * time.Second,
		NotifyBefore: 60 * time.Second,
	}, nil
}

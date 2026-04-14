package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DBURL           string `yaml:"db_url"`
	IntervalSec     int64  `yaml:"interval"`
	NotifyBeforeSec int64  `yaml:"notify_before"`
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
		DBURL:           dbURL,
		IntervalSec:     5,
		NotifyBeforeSec: 60,
	}, nil
}

package config

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	DBURL            string        `yaml:"db_url"`
	TrackingInterval time.Duration `yaml:"tracking_interval"`
	NotifyBefore     time.Duration `yaml:"notify_before"`
	NoLogout         bool          `yaml:"no_logout"`
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
		DBURL:            dbURL,
		TrackingInterval: 3 * time.Second,
		NotifyBefore:     1 * time.Minute,
	}, nil
}

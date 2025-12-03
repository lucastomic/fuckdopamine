package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the fuckdopamine configuration
type Config struct {
	BlockedSites []string `json:"blocked_sites"`
	LogFilePath  string   `json:"log_file_path"`
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	return "/etc/fuckdopamine"
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}

// GetStatsPath returns the full path to the stats file
func GetStatsPath() string {
	return "/var/lib/fuckdopamine/stats.json"
}

// Load loads the configuration from the config file
func Load() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves the configuration to the config file
func Save(cfg *Config) error {
	configDir := GetConfigDir()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := GetConfigPath()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		BlockedSites: []string{"example.com"},
		LogFilePath:  "/var/log/fuckdopamine/dns_requests.json",
	}
}

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads agent configuration from file
func LoadConfig(path string) (NodeConfig, error) {
	var config NodeConfig

	// Set defaults
	config.CoordinationURL = "http://localhost:3001"
	config.AuthToken = ""
	config.Provider = "docker"
	config.Heartbeat.Interval = 30 * time.Second
	config.Heartbeat.Timeout = 10 * time.Second
	config.Heartbeat.Retries = 3
	config.CommandTimeout = 5 * time.Minute
	config.RetryPolicy.MaxRetries = 3
	config.RetryPolicy.Backoff = time.Second
	config.RetryPolicy.MaxBackoff = 30 * time.Second
	config.OfflineMode = false
	config.CacheDir = "/tmp/nexus-agent"
	config.LogLevel = "info"

	// Load from file if exists
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return config, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return config, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Environment variable overrides
	if url := os.Getenv("NEXUS_COORD_URL"); url != "" {
		config.CoordinationURL = url
	}
	if token := os.Getenv("NEXUS_AUTH_TOKEN"); token != "" {
		config.AuthToken = token
	}
	if provider := os.Getenv("NEXUS_PROVIDER"); provider != "" {
		config.Provider = provider
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config NodeConfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default agent config path
func GetConfigPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "nexus", "agent.yaml")
	}

	return "/etc/nexus/agent.yaml"
}

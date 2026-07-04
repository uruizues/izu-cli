package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	v := viper.GetViper()

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, ".config", "izu-cli")
	configPath := filepath.Join(configDir, "config.yaml")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// Set defaults
	v.SetDefault("general.provider", "animekai")
	v.SetDefault("general.theme", "dark")
	v.SetDefault("general.language", "en")
	v.SetDefault("player.volume", 100)
	v.SetDefault("player.ipc_socket", "/tmp/izu-mpv-socket")
	v.SetDefault("download.enabled", true)
	v.SetDefault("download.quality", "best")
	v.SetDefault("download.concurrent", 3)
	v.SetDefault("storage.backend", "sqlite")
	v.SetDefault("storage.path", filepath.Join(configDir, "data"))

	// Try to read existing config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found, create default
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, err
		}
		if err := v.WriteConfigAs(configPath); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

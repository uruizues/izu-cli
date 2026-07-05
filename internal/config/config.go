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
	v.SetDefault("general.provider", "miruro")
	v.SetDefault("general.theme", "dark")
	v.SetDefault("general.language", "en")
	v.SetDefault("player.volume", 100)
	v.SetDefault("player.ipc_socket", "/tmp/izu-mpv-socket")
	v.SetDefault("download.enabled", true)
	v.SetDefault("download.quality", "best")
	v.SetDefault("download.concurrent", 3)
	v.SetDefault("storage.backend", "sqlite")
	v.SetDefault("storage.path", filepath.Join(configDir, "data"))

	// Provider defaults (disabled - using Consumet only)
	v.SetDefault("providers.animekai.enabled", false)
	v.SetDefault("providers.animekai.base_url", "https://animekai.to")
	v.SetDefault("providers.animekai.rate_limit", 30)
	v.SetDefault("providers.allanime.enabled", false)
	v.SetDefault("providers.allanime.api_url", "https://api.allanime.day/api")
	v.SetDefault("providers.allanime.referer", "https://youtu-chan.com")
	v.SetDefault("providers.allanime.rate_limit", 20)
	v.SetDefault("providers.animepahe.enabled", false)
	v.SetDefault("providers.animepahe.base_url", "https://animepahe.com")
	v.SetDefault("providers.animepahe.api_url", "https://animepahe.com/api.php")
	v.SetDefault("providers.animepahe.rate_limit", 20)

	// Consumet API provider defaults
	v.SetDefault("consumet.enabled", true)
	v.SetDefault("consumet.base_url", "http://localhost:3000")
	v.SetDefault("consumet.provider", "multi")
	
	v.SetDefault("providers.consumet_animekai.enabled", true)
	v.SetDefault("providers.consumet_animekai.base_url", "https://api.consumet.org")
	v.SetDefault("providers.consumet_animekai.rate_limit", 30)
	v.SetDefault("providers.consumet_hianime.enabled", true)
	v.SetDefault("providers.consumet_hianime.base_url", "https://api.consumet.org")
	v.SetDefault("providers.consumet_hianime.rate_limit", 30)
	v.SetDefault("providers.consumet_animepahe.enabled", false)
	v.SetDefault("providers.consumet_animepahe.base_url", "https://api.consumet.org")
	v.SetDefault("providers.consumet_animepahe.rate_limit", 20)
	v.SetDefault("providers.consumet_animesama.enabled", false)
	v.SetDefault("providers.consumet_animesama.base_url", "https://api.consumet.org")
	v.SetDefault("providers.consumet_animesama.rate_limit", 20)
	v.SetDefault("providers.consumet_animesaturn.enabled", false)
	v.SetDefault("providers.consumet_animesaturn.base_url", "https://api.consumet.org")
	v.SetDefault("providers.consumet_animesaturn.rate_limit", 20)
	v.SetDefault("providers.consumet_animeunity.enabled", false)
	v.SetDefault("providers.consumet_animeunity.base_url", "https://api.consumet.org")
	v.SetDefault("providers.consumet_animeunity.rate_limit", 20)
	v.SetDefault("providers.consumet_kickassanime.enabled", false)
	v.SetDefault("providers.consumet_kickassanime.base_url", "https://api.consumet.org")
	v.SetDefault("providers.consumet_kickassanime.rate_limit", 20)

	v.SetDefault("providers.miruro.enabled", false)
	v.SetDefault("providers.miruro.base_url", "http://localhost:8000")

	v.SetDefault("discord.enabled", false)
	v.SetDefault("discord.client_id", "")

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

package config

type Config struct {
	General   GeneralConfig   `mapstructure:"general"`
	Player    PlayerConfig    `mapstructure:"player"`
	Download  DownloadConfig  `mapstructure:"download"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Discord   DiscordConfig   `mapstructure:"discord"`
	Storage   StorageConfig   `mapstructure:"storage"`
}

type GeneralConfig struct {
	Provider string `mapstructure:"provider"`
	Theme    string `mapstructure:"theme"`
	Language string `mapstructure:"language"`
}

type PlayerConfig struct {
	Binary    string   `mapstructure:"binary"`
	Args      []string `mapstructure:"args"`
	IPCSocket string   `mapstructure:"ipc_socket"`
	Volume    int      `mapstructure:"volume"`
}

type DownloadConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Path       string `mapstructure:"path"`
	Quality    string `mapstructure:"quality"`
	Concurrent int    `mapstructure:"concurrent"`
}

type ProvidersConfig struct {
	AnimeKai  ProviderConfig `mapstructure:"animekai"`
	AllAnime  ProviderConfig `mapstructure:"allanime"`
	AnimePahe ProviderConfig `mapstructure:"animepahe"`
}

type ProviderConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	BaseURL   string `mapstructure:"base_url"`
	APIURL    string `mapstructure:"api_url"`
	Referer   string `mapstructure:"referer"`
	RateLimit int    `mapstructure:"rate_limit"`
}

type DiscordConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	ClientID string `mapstructure:"client_id"`
}

type StorageConfig struct {
	Backend string `mapstructure:"backend"`
	Path    string `mapstructure:"path"`
}

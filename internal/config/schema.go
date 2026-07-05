package config

type Config struct {
	General   GeneralConfig   `mapstructure:"general"`
	Player    PlayerConfig    `mapstructure:"player"`
	Download  DownloadConfig  `mapstructure:"download"`
	Providers ProvidersConfig `mapstructure:"providers"`
	Consumet  ConsumetConfig  `mapstructure:"consumet"`
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

type ConsumetConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	BaseURL  string `mapstructure:"base_url"`
	Provider string `mapstructure:"provider"`
}

type ProvidersConfig struct {
	ConsumetAnimeKai    ProviderConfig `mapstructure:"consumet_animekai"`
	ConsumetHiAnime     ProviderConfig `mapstructure:"consumet_hianime"`
	ConsumetAnimePahe   ProviderConfig `mapstructure:"consumet_animepahe"`
	ConsumetAnimeSama   ProviderConfig `mapstructure:"consumet_animesama"`
	ConsumetAnimeSaturn ProviderConfig `mapstructure:"consumet_animesaturn"`
	ConsumetAnimeUnity  ProviderConfig `mapstructure:"consumet_animeunity"`
	ConsumetKickass     ProviderConfig `mapstructure:"consumet_kickassanime"`
	Miruro              ProviderConfig `mapstructure:"miruro"`
	// Keep old providers for backward compatibility
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

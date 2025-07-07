package eventServer

type Config struct {
	Address string `mapstructure:"address"`
}

func (cfg *Config) InitDefaults() {
	if cfg.Address == "" {
		cfg.Address = ":8080"
	}
}

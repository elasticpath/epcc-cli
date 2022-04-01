package config

type Env struct {
	EPCC_API_BASE_URL  string `env:"EPCC_API_BASE_URL"`
	EPCC_CLIENT_ID     string `env:"EPCC_CLIENT_ID"`
	EPCC_CLIENT_SECRET string `env:"EPCC_CLIENT_SECRET"`
}

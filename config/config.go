package config

type Env struct {
	EPCC_API_BASE_URL                  string `env:"EPCC_API_BASE_URL"`
	EPCC_CLIENT_ID                     string `env:"EPCC_CLIENT_ID"`
	EPCC_CLIENT_SECRET                 string `env:"EPCC_CLIENT_SECRET"`
	EPCC_BETA_API_FEATURES             string `env:"EPCC_BETA_API_FEATURES"`
	EPCC_RATE_LIMIT                    uint16 `env:"EPCC_RATE_LIMIT"`
	EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES bool   `env:"EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES"`
}

var Envs = &Env{}

const DefaultUrl = "https://api.moltin.com"

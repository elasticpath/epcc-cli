package config

import "sync/atomic"

type Env struct {
	EPCC_API_BASE_URL                  string `env:"EPCC_API_BASE_URL"`
	EPCC_CLIENT_ID                     string `env:"EPCC_CLIENT_ID"`
	EPCC_CLIENT_SECRET                 string `env:"EPCC_CLIENT_SECRET"`
	EPCC_BETA_API_FEATURES             string `env:"EPCC_BETA_API_FEATURES"`
	EPCC_RATE_LIMIT                    uint16 `env:"EPCC_RATE_LIMIT"`
	EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES bool   `env:"EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES"`
	EPCC_RUNBOOK_DIRECTORY             string `env:"EPCC_RUNBOOK_DIRECTORY"`
}

var env = atomic.Pointer[Env]{}

func init() {
	SetEnv(&Env{})
}

func SetEnv(v *Env) {
	// Store a copy
	copyEnv := *v
	env.Store(&copyEnv)
}

func GetEnv() *Env {
	v := *env.Load()
	return &v
}

const DefaultUrl = "https://euwest.api.elasticpath.com"

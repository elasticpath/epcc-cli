package config

import "sync/atomic"

type Env struct {
	EPCC_API_BASE_URL                   string   `env:"EPCC_API_BASE_URL"`
	EPCC_CLIENT_ID                      string   `env:"EPCC_CLIENT_ID"`
	EPCC_CLIENT_SECRET                  string   `env:"EPCC_CLIENT_SECRET"`
	EPCC_BETA_API_FEATURES              string   `env:"EPCC_BETA_API_FEATURES"`
	EPCC_CLI_RATE_LIMIT                 uint16   `env:"EPCC_CLI_RATE_LIMIT"`
	EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES  bool     `env:"EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES"`
	EPCC_RUNBOOK_DIRECTORY              string   `env:"EPCC_RUNBOOK_DIRECTORY"`
	EPCC_DISABLE_LEGACY_RESOURCES       bool     `env:"EPCC_DISABLE_LEGACY_RESOURCES"`
	EPCC_CLI_DISABLE_RESOURCES          []string `env:"EPCC_CLI_DISABLE_RESOURCES" envSeparator:","`
	EPCC_CLI_DISABLE_TEMPLATE_EXECUTION bool     `env:"EPCC_CLI_DISABLE_TEMPLATE_EXECUTION"`
	EPCC_CLI_DISABLE_HTTP_LOGGING       bool     `env:"EPCC_CLI_DISABLE_HTTP_LOGGING"`
	EPCC_CLI_READ_ONLY                  bool     `env:"EPCC_CLI_READ_ONLY"`
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

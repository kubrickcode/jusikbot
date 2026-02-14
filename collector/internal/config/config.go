package config

import (
	env "github.com/caarlos0/env/v11"
)

// Env holds all environment-based configuration.
// Why caarlos0/env: zero-dependency, struct tags, required validation built-in.
// Why API keys are not required: each source validates its own keys at collection time,
// allowing `--target tiingo` to work without KIS keys and vice versa.
type Env struct {
	DatabaseURL  string `env:"DATABASE_URL,required,notEmpty"`
	KISAppKey    string `env:"KIS_APP_KEY"`
	KISAppSecret string `env:"KIS_APP_SECRET"`
	TiingoAPIKey string `env:"TIINGO_API_KEY"`
}

func LoadEnv() (Env, error) {
	return env.ParseAs[Env]()
}

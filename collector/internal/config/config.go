package config

import (
	env "github.com/caarlos0/env/v11"
)

// Env holds all environment-based configuration.
// Why caarlos0/env: zero-dependency, struct tags, required validation built-in.
type Env struct {
	DatabaseURL  string `env:"DATABASE_URL,required"`
	KISAppKey    string `env:"KIS_APP_KEY,required"`
	KISAppSecret string `env:"KIS_APP_SECRET,required"`
	TiingoAPIKey string `env:"TIINGO_API_KEY,required"`
}

func LoadEnv() (Env, error) {
	return env.ParseAs[Env]()
}

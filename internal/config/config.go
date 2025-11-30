package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port   string
	AppEnv string
	DB     DB
	APIKey string
}

type DB struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int64
}

func Load() Config {
	return Config{
		Port:   envString("PORT", "8080"),
		AppEnv: envString("APP_ENV", "dev"),
		DB: DB{
			DSN:          envString("DB_DSN", ""),
			MaxOpenConns: envInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns: envInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  envInt64("DB_MAX_LIFETIME", int64(30*time.Minute)),
		},
		APIKey: envString("API_KEY", "apitest"),
	}
}

// Helper with default fallback
func envString(key, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func envInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return i
}

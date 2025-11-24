package config

import "os"

type Config struct {
	Port   string
	AppEnv string
}

func Load() Config {
	return Config{
		Port:   getEnv("PORT", "8080"),
		AppEnv: getEnv("APP_ENV", "dev"),
	}
}

// Helper with default fallback
func getEnv(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}

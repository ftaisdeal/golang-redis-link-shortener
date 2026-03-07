package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	BaseURL            string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	CodeLength         int
	RateLimitPerMinute int
	ReservedKeywords   map[string]struct{}
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		BaseURL:            strings.TrimSuffix(getEnv("BASE_URL", "http://localhost:8080"), "/"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		CodeLength:         getEnvInt("CODE_LENGTH", 7),
		RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 60),
		ReservedKeywords: map[string]struct{}{
			"api":    {},
			"health": {},
		},
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

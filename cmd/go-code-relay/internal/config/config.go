package config

import (
	"os"
	"strconv"
)

type Config struct {
	Model            string
	APIKey           string
	BaseURL          string
	MaxTokens        int
	Temperature      float64
	MaxContextTokens int
	Provider         string
}

func (c *Config) SetDefaults() {
	c.Model = getEnv("CORECODER_MODEL", "gpt-4o")
	c.MaxTokens = intEnv("CORECODER_MAX_TOKENS", 4096)
	c.Temperature = floatEnv("CORECODER_TEMPERATURE", 0.0)
	c.MaxContextTokens = intEnv("CORECODER_MAX_CONTEXT", 128000)
	c.Provider = getEnv("CORECODER_PROVIDER", "openai")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func intEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func floatEnv(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

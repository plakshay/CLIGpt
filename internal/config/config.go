package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	GeminiAPIKey string
	Model        string
	Timeout      time.Duration
}

// Load reads .env (if present) and environment variables via Viper.
func Load() (*Config, error) {
	// Best-effort load of .env — ignore error if file is absent.
	_ = godotenv.Load()

	v := viper.New()
	v.AutomaticEnv()
	v.SetDefault("GEMINI_MODEL", "gemini-2.0-flash")
	v.SetDefault("REQUEST_TIMEOUT_SECONDS", 60)

	apiKey := v.GetString("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set (copy .env.example to .env and fill it in)")
	}

	return &Config{
		GeminiAPIKey: apiKey,
		Model:        v.GetString("GEMINI_MODEL"),
		Timeout:      time.Duration(v.GetInt("REQUEST_TIMEOUT_SECONDS")) * time.Second,
	}, nil
}
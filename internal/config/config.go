package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port                  string
	DBURL                 string
	JWTSecret             string
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	PaystackSecret        string
	PaystackBaseURL       string
	PaystackWebhookSecret string
}

// Load returns a Config populated from environment variables with reasonable defaults.
func Load() Config {
	cfg := Config{
		Port:                  getEnv("PORT", "8080"),
		DBURL:                 getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:     getEnv("GOOGLE_REDIRECT_URL", ""),
		PaystackSecret:        getEnv("PAYSTACK_SECRET_KEY", ""),
		PaystackBaseURL:       getEnv("PAYSTACK_BASE_URL", "https://api.paystack.co"),
		PaystackWebhookSecret: getEnv("PAYSTACK_WEBHOOK_SECRET", ""),
	}

	if cfg.DBURL == "" {
		log.Fatal("DATABASE_URL is required (e.g. postgres://user:pass@localhost:5432/wallet?sslmode=disable)")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.PaystackSecret == "" {
		log.Fatal("PAYSTACK_SECRET_KEY is required")
	}
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
		log.Fatal("GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URL are required")
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ParseExpiry converts the custom expiry strings (1H,1D,1M,1Y) into a duration.
func ParseExpiry(exp string) (time.Duration, error) {
	switch exp {
	case "1H":
		return time.Hour, nil
	case "1D":
		return 24 * time.Hour, nil
	case "1M":
		return 30 * 24 * time.Hour, nil
	case "1Y":
		return 365 * 24 * time.Hour, nil
	default:
		return 0, strconv.ErrSyntax
	}
}

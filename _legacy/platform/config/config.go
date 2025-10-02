package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	ServerHost string
	LogLevel   string
	LogFormat  string
	LogOutput  string

	DatabaseURL string

	WameowLogLevel string

	GlobalWebhookURL string
	WebhookSecret    string

	GlobalAPIKey string

	NodeEnv string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		_ = err
	}

	return &Config{
		Port:       getEnv("PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "http://localhost:8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		LogFormat:  getEnv("LOG_FORMAT", "console"),
		LogOutput:  getEnv("LOG_OUTPUT", "stdout"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/zpwoot?sslmode=disable"),

		WameowLogLevel: getEnv("WA_LOG_LEVEL", "INFO"),

		GlobalWebhookURL: getEnv("GLOBAL_WEBHOOK_URL", ""),
		WebhookSecret:    getEnv("WEBHOOK_SECRET", ""),

		GlobalAPIKey: getEnv("ZP_API_KEY", "a0b1125a0eb3364d98e2c49ec6f7d6ba"),

		NodeEnv: getEnv("NODE_ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) IsProduction() bool {
	return c.NodeEnv == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.NodeEnv == "development"
}

func (c *Config) IsTest() bool {
	return c.NodeEnv == "test"
}

func (c *Config) GetServerURL() string {
	return c.ServerHost
}

func (c *Config) HasWebhookSecret() bool {
	return c.WebhookSecret != ""
}

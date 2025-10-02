package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)


type Config struct {

	App AppConfig `json:"app"`


	Server ServerConfig `json:"server"`


	Log LogConfig `json:"log"`


	Database DatabaseConfig `json:"database"`


	WhatsApp WhatsAppConfig `json:"whatsapp"`


	Webhook WebhookConfig `json:"webhook"`


	Security SecurityConfig `json:"security"`


	Environment string `json:"environment"`
}


type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Debug   bool   `json:"debug"`
}


type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
	BaseURL      string `json:"base_url"`
}


type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Output string `json:"output"`
	Caller bool   `json:"caller"`
}


type DatabaseConfig struct {
	URL             string `json:"url"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
	MigrationsPath  string `json:"migrations_path"`
	AutoMigrate     bool   `json:"auto_migrate"`
}


type WhatsAppConfig struct {
	LogLevel     string `json:"log_level"`
	StoreDir     string `json:"store_dir"`
	MediaDir     string `json:"media_dir"`
	QRTimeout    int    `json:"qr_timeout"`
	PairTimeout  int    `json:"pair_timeout"`
	ReconnectMax int    `json:"reconnect_max"`
}


type WebhookConfig struct {
	GlobalURL    string `json:"global_url"`
	Secret       string `json:"secret"`
	Timeout      int    `json:"timeout"`
	RetryMax     int    `json:"retry_max"`
	RetryDelay   int    `json:"retry_delay"`
	VerifySSL    bool   `json:"verify_ssl"`
	UserAgent    string `json:"user_agent"`
}


type SecurityConfig struct {
	APIKey         string   `json:"api_key"`
	AllowedOrigins []string `json:"allowed_origins"`
	RateLimit      int      `json:"rate_limit"`
	RateLimitBurst int      `json:"rate_limit_burst"`
}


func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil {

	}

	config := &Config{
		App: AppConfig{
			Name:    getEnv("APP_NAME", "zpwoot"),
			Version: getEnv("APP_VERSION", "1.0.0"),
			Debug:   getEnvBool("APP_DEBUG", false),
		},

		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("PORT", 8080),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 120),
			BaseURL:      getEnv("SERVER_BASE_URL", "http://localhost:8080"),
		},

		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "console"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
			Caller: getEnvBool("LOG_CALLER", true),
		},

		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://zpwoot:zpwoot123@localhost:5432/zpwoot?sslmode=disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
			MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "internal/adapters/database/migrations"),
			AutoMigrate:     getEnvBool("DB_AUTO_MIGRATE", true),
		},

		WhatsApp: WhatsAppConfig{
			LogLevel:     getEnv("WA_LOG_LEVEL", "INFO"),
			StoreDir:     getEnv("WA_STORE_DIR", "./data/store"),
			MediaDir:     getEnv("WA_MEDIA_DIR", "./data/media"),
			QRTimeout:    getEnvInt("WA_QR_TIMEOUT", 120),
			PairTimeout:  getEnvInt("WA_PAIR_TIMEOUT", 60),
			ReconnectMax: getEnvInt("WA_RECONNECT_MAX", 5),
		},

		Webhook: WebhookConfig{
			GlobalURL:    getEnv("GLOBAL_WEBHOOK_URL", ""),
			Secret:       getEnv("WEBHOOK_SECRET", ""),
			Timeout:      getEnvInt("WEBHOOK_TIMEOUT", 30),
			RetryMax:     getEnvInt("WEBHOOK_RETRY_MAX", 3),
			RetryDelay:   getEnvInt("WEBHOOK_RETRY_DELAY", 5),
			VerifySSL:    getEnvBool("WEBHOOK_VERIFY_SSL", true),
			UserAgent:    getEnv("WEBHOOK_USER_AGENT", "zpwoot/1.0"),
		},

		Security: SecurityConfig{
			APIKey:         getEnv("ZP_API_KEY", "a0b1125a0eb3364d98e2c49ec6f7d6ba"),
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			RateLimit:      getEnvInt("RATE_LIMIT", 100),
			RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 10),
		},

		Environment: getEnv("NODE_ENV", "development"),
	}


	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}


func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	if c.Security.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	return nil
}


func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}


func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}


func (c *Config) IsTest() bool {
	return c.Environment == "test"
}


func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}


func (c *Config) HasWebhookSecret() bool {
	return c.Webhook.Secret != ""
}




func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}


func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}


func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

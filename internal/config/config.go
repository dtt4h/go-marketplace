package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	DevMode  bool
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Payments PaymentsConfig
	Redis    RedisConfig
	S3       S3Config
	SMTP     SMTPConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	CORSOrigins  string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type PaymentsConfig struct {
	GatewayURL    string
	GatewayKey    string
	WebhookSecret string
}

type RedisConfig struct {
	Host string
	Port int
	DB   int
}

type S3Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
}

type SMTPConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	FromEmail    string
	FromName     string
	Encryption   string // "tls", "starttls", or "none"
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:     getEnv("APP_ENV", "development"),
		DevMode: getEnv("APP_ENV", "development") == "development",
		Server: ServerConfig{
			Port:         getEnvInt("HTTP_PORT", 8080),
			ReadTimeout:  getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			CORSOrigins:  getEnv("CORS_ORIGINS", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "marketplace"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
			AccessTTL:  getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: getEnvDuration("JWT_REFRESH_TTL", 168*time.Hour),
		},
		Payments: PaymentsConfig{
			GatewayURL:    getEnv("PAYMENT_GATEWAY_URL", ""),
			GatewayKey:    getEnv("PAYMENT_GATEWAY_KEY", ""),
			WebhookSecret: getEnv("PAYMENTS_WEBHOOK_SECRET", ""),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnvInt("REDIS_PORT", 6379),
			DB:   getEnvInt("REDIS_DB", 0),
		},
		S3: S3Config{
			Endpoint:  getEnv("S3_ENDPOINT", ""),
			Region:    getEnv("S3_REGION", ""),
			AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""),
			Bucket:    getEnv("S3_BUCKET", ""),
		},
		SMTP: SMTPConfig{
			Host:       getEnv("SMTP_HOST", ""),
			Port:       getEnvInt("SMTP_PORT", 587),
			Username:   getEnv("SMTP_USERNAME", ""),
			Password:   getEnv("SMTP_PASSWORD", ""),
			FromEmail:  getEnv("SMTP_FROM_EMAIL", ""),
			FromName:   getEnv("SMTP_FROM_NAME", "Go Marketplace"),
			Encryption: getEnv("SMTP_ENCRYPTION", "starttls"),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

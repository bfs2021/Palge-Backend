package configs

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN string
	}
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}

	Limiter struct {
		Enabled         bool
		RPS             float64
		Burst           float64
		JanitorInterval time.Duration
		ClientExpiry    time.Duration
		TrustedProxies  []string
	}
}

func NewConfig(cfg *Config) {
	_ = godotenv.Load()

	cfg.Port = getEnvAsInt("PORT", 4000)
	cfg.Env = getEnv("ENV", "development")
	cfg.DB.DSN = getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5433/palge?sslmode=disable")

	cfg.SMTP.Host = getEnv("SMTP_HOST", "sandbox.smtp.mailtrap.io")
	cfg.SMTP.Port = getEnvAsInt("SMTP_PORT", 2525)
	cfg.SMTP.Username = getEnv("SMTP_USERNAME", "")
	cfg.SMTP.Password = getEnv("SMTP_PASSWORD", "")
	cfg.SMTP.Sender = getEnv("SMTP_SENDER", "Palge <no-reply@github.com/nongpal/Palge-Backend>")

	cfg.Limiter.Enabled = getEnvAsBool("RATE_LIMIT_ENABLED", true)
	cfg.Limiter.RPS = getEnvAsFloat("RATE_LIMIT_RPS", 2)
	cfg.Limiter.Burst = getEnvAsFloat("RATE_LIMIT_BURST", 4)
	cfg.Limiter.JanitorInterval = getEnvAsTime("RATE_LIMIT_JANITOR_INTERVAL", 1*time.Minute)
	cfg.Limiter.ClientExpiry = getEnvAsTime("RATE_LIMIT_CLIENT_EXPIRY", 10*time.Minute)
	cfg.Limiter.TrustedProxies = getEnvAsSlice("RATE_LIMIT_TRUSTED_PROXIES", []string{})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseFloat(value, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsTime(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if TimeValue, err := time.ParseDuration(value); err == nil {
			return TimeValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		sliceValue := strings.Split(value, ",")

		for i := range sliceValue {
			sliceValue[i] = strings.TrimSpace(sliceValue[i])
		}

		return sliceValue
	}
	return defaultValue
}

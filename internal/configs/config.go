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
	db   struct {
		dsn string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	rl struct {
		enabled          bool
		rps              float64
		burst            float64
		janitor_interval time.Duration
		client_expiry    time.Duration
		trustedProxies   []string
	}
}

func NewConfig(cfg *Config) {
	_ = godotenv.Load()

	cfg.Port = getEnvAsInt("PORT", 4000)
	cfg.Env = getEnv("ENV", "development")
	cfg.db.dsn = getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5433/palge?sslmode=disable")

	cfg.smtp.host = getEnv("SMTP_HOST", "sandbox.smtp.mailtrap.io")
	cfg.smtp.port = getEnvAsInt("SMTP_PORT", 2525)
	cfg.smtp.username = getEnv("SMTP_USERNAME", "")
	cfg.smtp.password = getEnv("SMTP_PASSWORD", "")
	cfg.smtp.sender = getEnv("SMTP_SENDER", "Palge <no-reply@github.com/nongpal/Palge-Backend>")

	cfg.rl.enabled = getEnvAsBool("RATE_LIMIT_ENABLED", true)
	cfg.rl.rps = getEnvAsFloat("RATE_LIMIT_RPS", 2)
	cfg.rl.burst = getEnvAsFloat("RATE_LIMIT_BURST", 4)
	cfg.rl.janitor_interval = getEnvAsTime("RATE_LIMIT_JANITOR_INTERVAL", 1*time.Minute)
	cfg.rl.client_expiry = getEnvAsTime("RATE_LIMIT_CLIENT_EXPIRY", 10*time.Minute)
	cfg.rl.trustedProxies = getEnvAsSlice("RATE_LIMIT_TRUSTED_PROXIES", []string{})
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

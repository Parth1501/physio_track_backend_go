package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds runtime configuration values.
type Config struct {
	Env             string
	Port            string
	DBUser          string
	DBPassword      string
	DBConnectString string
	TNSAdmin        string
	JWTSecret       string
	JWTIssuer       string
	JWTExpiry       time.Duration
}

// Load reads configuration from environment variables and .env (if present).
func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Env:             getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DBUser:          getEnv("DB_USER", ""),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		DBConnectString: getEnv("DB_CONNECT_STRING", ""),
		TNSAdmin:        getEnv("TNS_ADMIN", ""),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret"),
		JWTIssuer:       getEnv("JWT_ISSUER", "phsio-track"),
		JWTExpiry:       getEnvDuration("JWT_EXPIRY_MIN", 60) * time.Minute,
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBConnectString == "" || cfg.TNSAdmin == "" {
		log.Println("warning: database connection env vars incomplete (need DB_USER, DB_PASSWORD, DB_CONNECT_STRING, TNS_ADMIN)")
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvDuration(key string, defMinutes int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n)
		}
	}
	return time.Duration(defMinutes)
}

package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// defaultWhatsAppSignature closes WhatsApp messages sent to patients.
const defaultWhatsAppSignature = "Dr. Dency Singwala\n(MPT, COMT, CKT)"

// Config holds runtime configuration values.
type Config struct {
	Env             string
	Port            string
	DBUser          string
	DBPassword      string
	DBConnectString string
	TNSAdmin        string
	LogFile         string
	JWTSecret       string
	JWTIssuer       string
	JWTExpiry       time.Duration
	// WhatsAppSignature closes WhatsApp messages sent to patients.
	WhatsAppSignature string
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
		LogFile:         getEnv("LOG_FILE", ""),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret"),
		JWTIssuer:       getEnv("JWT_ISSUER", "phsio-track"),
		JWTExpiry:       getEnvDuration("JWT_EXPIRY_MIN", 60) * time.Minute,

		// A literal \n in the env value starts a new line in the signature.
		WhatsAppSignature: strings.ReplaceAll(getEnv("WHATSAPP_SIGNATURE", defaultWhatsAppSignature), `\n`, "\n"),
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

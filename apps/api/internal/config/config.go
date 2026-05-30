package config

import (
	"os"
)

// Config holds all application configuration read from environment variables.
// No hardcoded values — all settings come from the environment at runtime.
type Config struct {
	Port                     string
	SupabaseURL              string
	SupabaseAnonKey          string
	JWTSecret                string
	DatabaseURL              string
	IntegrationEncryptionKey string
}

// Load reads configuration from environment variables.
// Call this once at startup, after loading .env with godotenv.
func Load() *Config {
	return &Config{
		Port:                     getEnv("PORT", "8080"),
		SupabaseURL:              getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:          getEnv("SUPABASE_ANON_KEY", ""),
		JWTSecret:                getEnv("JWT_SECRET", ""),
		DatabaseURL:              getEnv("DATABASE_URL", ""),
		IntegrationEncryptionKey: getEnv("INTEGRATION_ENCRYPTION_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

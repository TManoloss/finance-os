package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config armazena todas as configurações do sistema carregadas do ambiente.
type Config struct {
	DatabaseURL         string
	JWTSecret           string
	JWTRefreshSecret    string
	PluggyClientID      string
	PluggyClientSecret  string
	AgentsServiceURL    string
	Port                string
	CORSOrigins         string
}

// Load carrega as configurações de variáveis de ambiente.
func Load() (*Config, error) {
	// Carrega .env se existir, mas não falha se não existir (ex: produção)
	_ = godotenv.Load()

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://finance:finance123@localhost:5432/financedb"),
		JWTSecret:          getEnv("JWT_SECRET", "secret"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "refresh_secret"),
		PluggyClientID:     getEnv("PLUGGY_CLIENT_ID", ""),
		PluggyClientSecret: getEnv("PLUGGY_CLIENT_SECRET", ""),
		AgentsServiceURL:   getEnv("AGENTS_SERVICE_URL", "http://localhost:8000"),
		Port:               getEnv("PORT", "8080"),
		CORSOrigins:        getEnv("CORS_ORIGINS", "*"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

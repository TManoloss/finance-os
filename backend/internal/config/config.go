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
	EncryptionKey       string
	PluggyClientID      string
	PluggyClientSecret  string
	AgentsServiceURL    string
	SelfURL             string
	SyncSecret          string
	Port                string
	CORSOrigins         string
}

// Load carrega as configurações de variáveis de ambiente.
func Load() (*Config, error) {
	// Carrega .env se existir no diretório atual ou no pai (caso esteja rodando dentro de backend/)
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://finance:finance123@localhost:5432/financedb"),
		JWTSecret:          getEnv("JWT_SECRET", "secret"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "refresh_secret"),
		EncryptionKey:      getEnv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef"), // 32 bytes default for AES-256
		PluggyClientID:     getEnv("PLUGGY_CLIENT_ID", ""),
		PluggyClientSecret: getEnv("PLUGGY_CLIENT_SECRET", ""),
		AgentsServiceURL:   getEnv("AGENTS_SERVICE_URL", "http://localhost:8000"),
		SelfURL:            getEnv("SELF_URL", ""),
		SyncSecret:         getEnv("SYNC_SECRET", ""),
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

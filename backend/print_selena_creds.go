package main

import (
	"context"
	"fmt"
	"log"

	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := "postgresql://postgres.kuimkqgvotrkzlrhdqhp:Quimeras40**@aws-1-us-east-1.pooler.supabase.com:6543/postgres"

	ctx := context.Background()
	connConfig, err := pgx.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Erro no parse da URL: %v", err)
	}
	connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco de produção: %v", err)
	}
	defer conn.Close(ctx)

	var clientID, encryptedSecret string
	err = conn.QueryRow(ctx, "SELECT pluggy_client_id, pluggy_client_secret_encrypted FROM users WHERE email = 'selenapereira42@gmail.com'").Scan(&clientID, &encryptedSecret)
	if err != nil {
		log.Fatalf("Erro ao buscar Selena: %v", err)
	}

	encService, err := service.NewEncryptionService("MnwcJ6VXteWIF/Qca83Keia5qoRpAxHpJnDD7+aTN+M=")
	if err != nil {
		log.Fatalf("Erro ao iniciar EncryptionService: %v", err)
	}

	clientSecret, err := encService.Decrypt(encryptedSecret)
	if err != nil {
		log.Fatalf("Erro ao descriptografar secret: %v", err)
	}

	fmt.Println("=== VERIFICAÇÃO DE CREDENCIAIS DA SELENA ===")
	fmt.Printf("Client ID: %s\n", clientID)
	
	// Mascarar secret para mostrar se é um valor válido
	masked := ""
	if len(clientSecret) > 8 {
		masked = clientSecret[:4] + "..." + clientSecret[len(clientSecret)-4:]
	} else {
		masked = "muito curto: " + clientSecret
	}
	fmt.Printf("Client Secret Decriptado (Mascarado): %s (Comprimento: %d)\n", masked, len(clientSecret))
}

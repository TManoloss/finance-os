package main

import (
	"context"
	"fmt"
	"log"

	"github.com/finance-os/backend/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}
	defer conn.Close(ctx)

	email := "manoelfelipe90@gmail.com"
	
	// 1. Verificar estado atual
	var clientID, secret string
	err = conn.QueryRow(ctx, "SELECT COALESCE(pluggy_client_id, 'NULL'), COALESCE(pluggy_client_secret_encrypted, 'NULL') FROM users WHERE email = $1", email).Scan(&clientID, &secret)
	if err != nil {
		log.Fatalf("Erro ao buscar usuário: %v", err)
	}

	fmt.Printf("Estado atual para %s:\n", email)
	fmt.Printf("  pluggy_client_id: %s\n", clientID)
	fmt.Printf("  pluggy_client_secret: %s\n", secret)

	// 2. Limpar se necessário
	if clientID != "NULL" || secret != "NULL" {
		fmt.Println("\nLimpando credenciais residuais...")
		_, err = conn.Exec(ctx, "UPDATE users SET pluggy_client_id = NULL, pluggy_client_secret_encrypted = NULL WHERE email = $1", email)
		if err != nil {
			log.Fatalf("Erro ao limpar credenciais: %v", err)
		}
		fmt.Println("✓ Credenciais removidas com sucesso.")
	} else {
		fmt.Println("\nO usuário já está limpo no banco de dados.")
	}
}

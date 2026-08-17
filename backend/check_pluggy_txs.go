package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro config: %v", err)
	}

	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Erro bd: %v", err)
	}
	defer dbPool.Close()

	var userID, itemID, accountID string
	err = dbPool.QueryRow(context.Background(), "SELECT user_id, pluggy_item_id, pluggy_account_id FROM connected_accounts LIMIT 1").Scan(&userID, &itemID, &accountID)
	if err != nil {
		log.Fatalf("Nenhuma conta conectada: %v", err)
	}

	var clientID, clientSecretEncrypted string
	err = dbPool.QueryRow(context.Background(), "SELECT pluggy_client_id, pluggy_client_secret_encrypted FROM users WHERE id = $1", userID).Scan(&clientID, &clientSecretEncrypted)
	if err != nil {
		log.Fatalf("Erro ao buscar credenciais do usuario: %v", err)
	}

	encryptionService, _ := service.NewEncryptionService(cfg.EncryptionKey)
	clientSecret, _ := encryptionService.Decrypt(clientSecretEncrypted)

	client := pluggy.NewClient(clientID, clientSecret)

	to := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	from := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	
	fmt.Printf("Fetching from %s to %s for account %s (user %s)...\n", from, to, accountID, userID)
	txs, err := client.GetTransactions(accountID, from, to)
	if err != nil {
		log.Fatalf("Erro API: %v", err)
	}

	fmt.Printf("Found %d transactions.\n", len(txs))
	for i, tx := range txs {
		if i < 15 {
			fmt.Printf("- %s | %s | %.2f | %s\n", tx.Date, tx.Description, tx.Amount, tx.Status)
		}
	}
}

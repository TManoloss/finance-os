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
		log.Fatal(err)
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}
	defer db.Close()

	userID := "eb59b2b0-8325-4de2-87f6-fcaac68a2b33" // Sabrina

	// 1. Buscar credenciais
	var pluggyClientID, encryptedSecret string
	err = db.QueryRow(ctx, "SELECT COALESCE(pluggy_client_id, ''), COALESCE(pluggy_client_secret_encrypted, '') FROM users WHERE id = $1", userID).Scan(&pluggyClientID, &encryptedSecret)
	if err != nil {
		log.Fatalf("Erro ao buscar credenciais: %v", err)
	}

	fmt.Printf("Credenciais encriptadas obtidas:\n  Client ID: %s\n  Secret Encrypted: %s\n", pluggyClientID, encryptedSecret)

	if pluggyClientID == "" || encryptedSecret == "" {
		log.Fatal("Usuário não tem credenciais configuradas!")
	}

	// 2. Decriptar
	encService, err := service.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Erro ao iniciar EncryptionService: %v", err)
	}

	decryptedSecret, err := encService.Decrypt(encryptedSecret)
	if err != nil {
		log.Fatalf("Erro ao decriptar: %v", err)
	}

	fmt.Printf("Decriptado com sucesso! Client Secret: %s\n", decryptedSecret)

	// 3. Testar conexão com Pluggy
	fmt.Println("\n=== Testando conexão com Pluggy ===")
	pluggyClient := pluggy.NewClient(pluggyClientID, decryptedSecret)
	
	// Tentar obter as contas dela no Pluggy usando um dos itens conhecidos
	// Item IDs de Sabrina: b0f3b86f-6d76-4367-9148-f8eee586872e, 54577609-e445-4466-947f-b9763e4cd726
	itemIDs := []string{"b0f3b86f-6d76-4367-9148-f8eee586872e", "54577609-e445-4466-947f-b9763e4cd726"}
	for _, itemID := range itemIDs {
		fmt.Printf("\n--- Consultando Item %s ---\n", itemID)
		item, err := pluggyClient.GetItem(itemID)
		if err != nil {
			fmt.Printf("  Erro ao obter Item: %v\n", err)
			continue
		}
		fmt.Printf("  Item ID: %s | Status: %s | Connector: %d\n", item.ID, item.Status, item.ConnectorID)

		// Buscar contas
		accounts, err := pluggyClient.GetAccounts(itemID)
		if err != nil {
			fmt.Printf("  Erro ao obter Contas do Item: %v\n", err)
			continue
		}
		for _, acc := range accounts {
			fmt.Printf("    └─ Conta: %s | Tipo: %s | Saldo: %.2f %s\n", acc.MarketingName, acc.Type, acc.Balance, acc.CurrencyCode)
			
			// Testar busca de transações dos últimos 5 dias
			to := time.Now().Format("2006-01-02")
			from := time.Now().AddDate(0, 0, -5).Format("2006-01-02")
			txs, err := pluggyClient.GetTransactions(acc.ID, from, to)
			if err != nil {
				fmt.Printf("      Erro ao obter Transações: %v\n", err)
			} else {
				fmt.Printf("      Transações obtidas para o período (%s a %s): %d\n", from, to, len(txs))
			}
		}
	}

	// 4. Executar Sync completo para ela
	fmt.Println("\n=== Executando SyncUserAccounts completo ===")
	// Criar instâncias dos serviços dependentes
	classifier := service.NewClassifierService(db, cfg)
	feed := service.NewFeedService(db)
	installments := service.NewInstallmentsService(db)
	syncService := service.NewSyncService(db, installments, classifier, feed)

	totalSaved, err := syncService.SyncUserAccounts(ctx, userID, pluggyClient)
	if err != nil {
		log.Fatalf("Erro ao rodar SyncUserAccounts: %v", err)
	}
	fmt.Printf("\nSync finalizado! Total de novas transações importadas: %d\n", totalSaved)
}

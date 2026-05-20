package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	fmt.Println("=== BUSCANDO POR SELENA OU TODOS OS USUÁRIOS ===")
	rows, err := conn.Query(ctx, `
		SELECT id, name, email, created_at, COALESCE(pluggy_client_id, 'NULL')
		FROM users 
		WHERE name ILIKE '%selena%' OR email ILIKE '%selena%'
	`)
	if err != nil {
		log.Fatalf("Erro ao buscar Selena: %v", err)
	}
	defer rows.Close()

	found := false
	var selenaID string
	for rows.Next() {
		found = true
		var id, name, email, pluggyID string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &email, &createdAt, &pluggyID); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[ENCONTRADA] User: %s (%s) | ID: %s | Criado em: %s | Pluggy Client ID: %s\n", name, email, id, createdAt.Format("02/01/2006 15:04:05"), pluggyID)
		selenaID = id
	}
	
	if !found {
		fmt.Println("Nenhum usuário com o nome ou e-mail contendo 'selena' foi encontrado!")
		fmt.Println("Listando todos os usuários do banco:")
		allRows, err := conn.Query(ctx, `SELECT id, name, email, created_at, COALESCE(pluggy_client_id, 'NULL') FROM users ORDER BY created_at DESC`)
		if err == nil {
			defer allRows.Close()
			for allRows.Next() {
				var id, name, email, pluggyID string
				var createdAt time.Time
				allRows.Scan(&id, &name, &email, &createdAt, &pluggyID)
				fmt.Printf("  User: %s (%s) | ID: %s | Criado em: %s | Pluggy ID: %s\n", name, email, id, createdAt.Format("02/01/2006 15:04:05"), pluggyID)
			}
		}
		return
	}

	// Se encontrou Selena, listar as contas dela
	fmt.Printf("\n=== CONTAS DE SELENA (ID: %s) ===\n", selenaID)
	accRows, err := conn.Query(ctx, `
		SELECT id, COALESCE(pluggy_item_id, 'NULL'), COALESCE(institution_name, 'NULL'), account_type, balance, last_synced_at 
		FROM connected_accounts 
		WHERE user_id = $1
	`, selenaID)
	if err != nil {
		log.Fatalf("Erro ao buscar contas: %v", err)
	}
	defer accRows.Close()

	hasAccounts := false
	for accRows.Next() {
		hasAccounts = true
		var id, itemID, instName, accType string
		var balance float64
		var lastSynced *time.Time
		if err := accRows.Scan(&id, &itemID, &instName, &accType, &balance, &lastSynced); err != nil {
			log.Fatal(err)
		}
		syncedStr := "Nunca"
		if lastSynced != nil {
			syncedStr = lastSynced.Format("02/01/2006 15:04:05")
		}

		// Contar transações
		var txCount int
		err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM transactions WHERE account_id = $1", id).Scan(&txCount)
		if err != nil {
			txCount = -1
		}

		fmt.Printf("  └─ Conta: %s (%s) | Tipo: %s | Saldo: %.2f | Sinc: %s | Transações: %d\n", 
			instName, itemID, accType, balance, syncedStr, txCount)
	}

	if !hasAccounts {
		fmt.Println("  (Nenhuma conta conectada para Selena)")
	}
}

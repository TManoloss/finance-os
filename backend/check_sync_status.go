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

	fmt.Println("=== 1. BUSCANDO POR SELENA OU TODOS OS USUÁRIOS ===")
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
	for rows.Next() {
		found = true
		var id, name, email, pluggyID string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &email, &createdAt, &pluggyID); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[ENCONTRADA] User: %s (%s) | ID: %s | Criado em: %s | Pluggy Client ID: %s\n", name, email, id, createdAt.Format("02/01/2006 15:04:05"), pluggyID)
	}
	
	if !found {
		fmt.Println("Nenhum usuário com o nome ou e-mail contendo 'selena' foi encontrado!")
		fmt.Println("Listando todos os usuários novamente:")
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
	}


	fmt.Println("\n=== 2. CONTAS CONECTADAS POR USUÁRIO ===")
	for _, u := range users {
		fmt.Printf("\n--- Contas de %s (%s) ---\n", u.Name, u.Email)
		
		type Account struct {
			ID              string
			ItemID          string
			InstitutionName string
			AccountType     string
			Balance         float64
			LastSynced      *time.Time
		}
		var accounts []Account
		
		accRows, err := conn.Query(ctx, `
			SELECT id, COALESCE(pluggy_item_id, 'NULL'), COALESCE(institution_name, 'NULL'), account_type, balance, last_synced_at 
			FROM connected_accounts 
			WHERE user_id = $1
		`, u.ID)
		if err != nil {
			log.Printf("Erro ao buscar contas: %v", err)
			continue
		}
		
		for accRows.Next() {
			var acc Account
			if err := accRows.Scan(&acc.ID, &acc.ItemID, &acc.InstitutionName, &acc.AccountType, &acc.Balance, &acc.LastSynced); err != nil {
				log.Fatal(err)
			}
			accounts = append(accounts, acc)
		}
		if len(accounts) == 0 {
			fmt.Println("  (Nenhuma conta conectada)")
			continue
		}

		for _, acc := range accounts {
			syncedStr := "Nunca"
			if acc.LastSynced != nil {
				syncedStr = acc.LastSynced.Format("02/01/2006 15:04:05")
			}
			fmt.Printf("  └─ Conta: %s (%s) | Tipo: %s | Saldo: %.2f | Sinc: %s\n", 
				acc.InstitutionName, acc.ItemID, acc.AccountType, acc.Balance, syncedStr)
		}
	}

	fmt.Println("\n=== 3. RECENTES SYNC LOGS ===")
	logRows, err := conn.Query(ctx, `
		SELECT id, triggered_by, synced_users, transactions_imported, errors_count, errors_detail, duration_ms, started_at, finished_at 
		FROM sync_logs 
		ORDER BY started_at DESC 
		LIMIT 10
	`)
	if err == nil {
		defer logRows.Close()
		for logRows.Next() {
			var id, triggeredBy string
			var syncedUsers, txImported, errCount, durationMs int
			var errDetail []byte
			var startedAt, finishedAt time.Time
			if err := logRows.Scan(&id, &triggeredBy, &syncedUsers, &txImported, &errCount, &errDetail, &durationMs, &startedAt, &finishedAt); err != nil {
				log.Printf("Erro ao escanear sync log: %v", err)
				continue
			}
			fmt.Printf("Log ID: %s | Trig: %s | Users: %d | Txs: %d | Erros: %d | Duração: %dms | Iniciado: %s | Detalhes: %s\n", 
				id, triggeredBy, syncedUsers, txImported, errCount, durationMs, startedAt.Format("02/01/2006 15:04:05"), string(errDetail))
		}
	} else {
		fmt.Printf("Erro ao buscar sync_logs: %v (a tabela pode não existir)\n", err)
	}
}




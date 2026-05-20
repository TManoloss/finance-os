package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	// URL explícita do Supabase em produção
	dbURL := "postgresql://postgres.kuimkqgvotrkzlrhdqhp:Quimeras40**@aws-1-us-east-1.pooler.supabase.com:6543/postgres"

	ctx := context.Background()
	connConfig, err := pgx.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Erro ao fazer parse da URL do banco: %v", err)
	}
	connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco de produção: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("=== CONECTADO AO SUPABASE DE PRODUÇÃO EXPLICITAMENTE ===")

	// 1. Buscar Selena
	emailQuery := "selenapereira42@gmail.com"
	var selenaID, name, email string
	var createdAt time.Time
	var pluggyClientID *string
	err = conn.QueryRow(ctx, `
		SELECT id, name, email, created_at, pluggy_client_id 
		FROM users 
		WHERE email = $1`, emailQuery).Scan(&selenaID, &name, &email, &createdAt, &pluggyClientID)

	if err != nil {
		fmt.Printf("Erro ao buscar a usuária %s: %v\n", emailQuery, err)
		
		// Listar todos os usuários da produção
		fmt.Println("\nListando todos os usuários no banco de produção:")
		rows, err := conn.Query(ctx, "SELECT id, name, email, created_at FROM users ORDER BY created_at DESC")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, n, e string
				var c time.Time
				rows.Scan(&id, &n, &e, &c)
				fmt.Printf("  User: %s (%s) | ID: %s | Criado em: %s\n", n, e, id, c.Format("02/01/2006 15:04:05"))
			}
		}
		return
	}

	pluggyStr := "NULL"
	if pluggyClientID != nil {
		pluggyStr = *pluggyClientID
	}
	fmt.Printf("\n[SUCESSO] Usuária encontrada:\n  Nome: %s\n  E-mail: %s\n  ID: %s\n  Criada em: %s\n  Pluggy Client ID: %s\n", 
		name, email, selenaID, createdAt.Format("02/01/2006 15:04:05"), pluggyStr)

	// 2. Buscar contas de Selena
	fmt.Println("\n=== CONTAS CONECTADAS DE SELENA ===")
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

	// 3. Buscar logs de sincronização recentes para Selena
	fmt.Println("\n=== LOGS DE SINCRONIZAÇÃO RECENTES (PRODUÇÃO) ===")
	logRows, err := conn.Query(ctx, `
		SELECT id, triggered_by, transactions_imported, errors_count, errors_detail, started_at, finished_at
		FROM sync_logs
		WHERE errors_detail::text LIKE '%' || $1 || '%'
		   OR triggered_by = 'manual' -- Listar todos os manuais recentes para ver se algum pertence a ela sem logar erro
		ORDER BY started_at DESC LIMIT 10
	`, selenaID)
	if err != nil {
		fmt.Printf("Erro ao buscar logs de sync: %v\n", err)
		return
	}
	defer logRows.Close()

	hasLogs := false
	for logRows.Next() {
		hasLogs = true
		var id string
		var triggeredBy string
		var txImported, errCount int
		var errDetail string
		var startedAt, finishedAt time.Time
		if err := logRows.Scan(&id, &triggeredBy, &txImported, &errCount, &errDetail, &startedAt, &finishedAt); err != nil {
			log.Fatal(err)
		}
		duration := finishedAt.Sub(startedAt).Milliseconds()
		fmt.Printf("  Log: %s | Origem: %s | Importadas: %d | Erros: %d | Duração: %dms | Iniciado: %s\n", 
			id[:8], triggeredBy, txImported, errCount, duration, startedAt.Format("02/01/2006 15:04:05"))
		if errCount > 0 {
			fmt.Printf("    └─ Detalhes: %s\n", errDetail)
		}
	}
	if !hasLogs {
		fmt.Println("  (Nenhum log de sincronização encontrado)")
	}
}

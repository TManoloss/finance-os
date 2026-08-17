package main

import (
	"context"
	"fmt"
	"log"
	"github.com/finance-os/backend/internal/config"
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

	rows, err := dbPool.Query(context.Background(), "SELECT started_at, triggered_by, duration_ms, synced_users, transactions_imported, errors_count, errors_detail FROM sync_logs ORDER BY started_at DESC LIMIT 3")
	if err != nil {
		log.Fatalf("Erro query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var startedAt, triggeredBy, durationMs, syncedUsers, txImported, errorsCount interface{}
		var errorsDetail string
		if err := rows.Scan(&startedAt, &triggeredBy, &durationMs, &syncedUsers, &txImported, &errorsCount, &errorsDetail); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Started: %v, By: %v, Dur: %v, Users: %v, Tx: %v, ErrCnt: %v, Details: %s\n", startedAt, triggeredBy, durationMs, syncedUsers, txImported, errorsCount, errorsDetail)
	}
}

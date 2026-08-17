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

	res, err := dbPool.Exec(context.Background(), "UPDATE connected_accounts SET last_synced_at = NOW() - INTERVAL '1 hour'")
	if err != nil {
		log.Fatalf("Erro update: %v", err)
	}
	fmt.Printf("Contas resetadas: %v\n", res.RowsAffected())
}

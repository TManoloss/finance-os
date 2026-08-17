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

	rows, err := dbPool.Query(context.Background(), "SELECT id, user_id, institution_name, balance FROM connected_accounts")
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, uid, inst string
		var bal float64
		if err := rows.Scan(&id, &uid, &inst, &bal); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Acc: %s | User: %s | Inst: %s | Bal: %.2f\n", id, uid, inst, bal)
	}
}

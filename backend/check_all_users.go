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

	rows, err := dbPool.Query(context.Background(), "SELECT id, name, email, COALESCE(pluggy_client_id, '') FROM users")
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, email, clientID string
		if err := rows.Scan(&id, &name, &email, &clientID); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User: %s | Email: %s | ClientID: %s\n", name, email, clientID)
	}
}

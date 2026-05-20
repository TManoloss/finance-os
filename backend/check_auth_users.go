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

	fmt.Println("=== BUSCANDO NA TABELA auth.users ===")
	// Alguns projetos Supabase têm uma tabela auth.users na schema 'auth'
	var authCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM auth.users").Scan(&authCount)
	if err != nil {
		fmt.Printf("Erro ou tabela auth.users não acessível: %v\n", err)
	} else {
		fmt.Printf("Tabela auth.users tem %d usuários.\n", authCount)
		rows, err := conn.Query(ctx, "SELECT id, email, created_at FROM auth.users")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, email string
				var createdAt interface{}
				rows.Scan(&id, &email, &createdAt)
				fmt.Printf("  Auth User - ID: %s | E-mail: %s | Criado: %v\n", id, email, createdAt)
			}
		}
	}

	fmt.Println("\n=== BUSCANDO NA TABELA public.users ===")
	rows, err := conn.Query(ctx, "SELECT id, name, email, created_at FROM public.users")
	if err != nil {
		log.Fatalf("Erro ao listar public.users: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, email string
		var createdAt interface{}
		rows.Scan(&id, &name, &email, &createdAt)
		fmt.Printf("  Public User - ID: %s | Nome: %s | E-mail: %s | Criado: %v\n", id, name, email, createdAt)
	}
}

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

	fmt.Println("=== 1. TABELAS NO BANCO ===")
	rows, err := conn.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`)
	if err != nil {
		log.Fatalf("Erro ao listar tabelas: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}

	for _, t := range tables {
		var count int
		conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", t)).Scan(&count)
		fmt.Printf("Tabela: %s | Linhas: %d\n", t, count)
	}

	fmt.Println("\n=== 2. TODOS OS USUÁRIOS DETALHADOS ===")
	userRows, err := conn.Query(ctx, `SELECT id, name, email, created_at FROM users`)
	if err != nil {
		log.Fatalf("Erro ao listar usuários: %v", err)
	}
	defer userRows.Close()

	for userRows.Next() {
		var id, name, email string
		var createdAt interface{}
		userRows.Scan(&id, &name, &email, &createdAt)
		fmt.Printf("ID: %s | Nome: %s | E-mail: %s | Criado em: %v\n", id, name, email, createdAt)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/repository"
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
	repo := repository.NewTransactionRepository(db)

	// Teste 1: Datas vazias (IsZero)
	fmt.Println("=== TESTE 1: DATAS ZERADAS ===")
	summary, err := repo.GetSummary(ctx, userID, time.Time{}, time.Time{})
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Printf("Total Gasto: %.2f\n", summary.TotalSpent)
		fmt.Printf("Total Recebido: %.2f\n", summary.TotalReceived)
		fmt.Printf("Saldo Checking: %.2f\n", summary.CheckingBalance)
		fmt.Printf("Qtd por Categoria: %d\n", len(summary.ByCategory))
	}

	// Teste 2: Mês atual (Maio 2026)
	fmt.Println("\n=== TESTE 2: MAIO 2026 ===")
	fromDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC)
	summary2, err := repo.GetSummary(ctx, userID, fromDate, toDate)
	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Printf("Total Gasto: %.2f\n", summary2.TotalSpent)
		fmt.Printf("Total Recebido: %.2f\n", summary2.TotalReceived)
		fmt.Printf("Saldo Checking: %.2f\n", summary2.CheckingBalance)
		fmt.Printf("Qtd por Categoria: %d\n", len(summary2.ByCategory))
	}
}

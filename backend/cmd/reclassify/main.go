package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	dbConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	db, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	classifier := service.NewClassifierService(db, cfg)

	// Busca transações "Outros"
	rows, err := db.Query(ctx, `
		SELECT t.id, t.description, t.merchant_name, t.amount, t.direction, acc.user_id
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		JOIN categories c ON t.category_id = c.id
		WHERE c.name = 'Outros'
		LIMIT 100
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Iniciando reclassificação em massa...")

	for rows.Next() {
		var tx struct {
			ID           string
			Description  string
			MerchantName *string
			Amount       float64
			Direction    string
			UserID       string
		}
		if err := rows.Scan(&tx.ID, &tx.Description, &tx.MerchantName, &tx.Amount, &tx.Direction, &tx.UserID); err != nil {
			log.Printf("Erro ao ler linha: %v", err)
			continue
		}

		merchant := tx.Description
		if tx.MerchantName != nil && *tx.MerchantName != "" {
			merchant = *tx.MerchantName
		}

		fmt.Printf("Classificando: %s (R$ %.2f)...\n", tx.Description, tx.Amount)
		
		categoryID, err := classifier.Classify(ctx, tx.UserID, merchant, tx.Description, tx.Amount, tx.Direction)
		if err != nil {
			log.Printf("Erro ao classificar %s: %v", tx.ID, err)
			continue
		}

		// Busca nome da categoria
		var catName string
		db.QueryRow(ctx, "SELECT name FROM categories WHERE id = $1", categoryID).Scan(&catName)

		fmt.Printf("  -> IA Sugeriu: %s (ID: %s)\n", catName, categoryID)

		// Atualiza se for diferente de Outros
		if catName != "Outros" {
			_, err = db.Exec(ctx, "UPDATE transactions SET category_id = $1 WHERE id = $2", categoryID, tx.ID)
			if err != nil {
				log.Printf("Erro ao atualizar %s: %v", tx.ID, err)
			} else {
				fmt.Printf("  -> Sucesso: %s\n", catName)
			}
		} else {
			fmt.Println("  -> Mantido como Outros.")
		}

		// Pequena pausa para respeitar o log do backend (embora o classifier já tenha sleep)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Reclassificação concluída!")
}

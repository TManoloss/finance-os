package main

import (
	"context"
	"log"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, _ := config.Load()
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	encryptionService, _ := service.NewEncryptionService(cfg.EncryptionKey)
	
	installmentService := service.NewInstallmentsService(db)
	classifierService := service.NewClassifierService(db, cfg)
	feedService := service.NewFeedService(db)
	syncService := service.NewSyncService(db, installmentService, classifierService, feedService)

	// Fetch all users with connected accounts
	rows, err := db.Query(context.Background(), "SELECT DISTINCT user_id FROM connected_accounts")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var uid string
		rows.Scan(&uid)
		userIDs = append(userIDs, uid)
	}

	for _, userID := range userIDs {
		// Pega credenciais do usuário
		user, _ := userRepo.FindByID(context.Background(), userID)
		
		clientID := user.PluggyClientID
		clientSecret := ""
		if user.PluggyClientSecretEncrypted != "" {
			clientSecret, _ = encryptionService.Decrypt(user.PluggyClientSecretEncrypted)
		}
		
		if clientID == "" || clientSecret == "" {
			clientID = cfg.PluggyClientID
			clientSecret = cfg.PluggyClientSecret
		}

		pluggyClient := pluggy.NewClient(clientID, clientSecret)
		
		log.Printf("Sincronizando user: %s", userID)
		count, err := syncService.SyncUserAccounts(context.Background(), userID, pluggyClient)
		if err != nil {
			log.Printf("Erro: %v", err)
		} else {
			log.Printf("Sucesso. %d transacoes importadas", count)
		}
	}
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/jobs"
	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/router"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// 2. Conectar ao PostgreSQL
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer dbPool.Close()

	// Testar conexão
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Erro ao pingar o banco de dados: %v", err)
	}
	log.Println("Conectado ao PostgreSQL com sucesso!")

	// 3. Inicializar Echo
	e := echo.New()

	// 4. Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{cfg.CORSOrigins},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	// 5. Inicializar Scheduler
	userRepo := repository.NewUserRepository(dbPool)
	encryptionService, err := service.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Erro ao inicializar encryption service: %v", err)
	}
	
	installmentService := service.NewInstallmentsService(dbPool)
	classifierService := service.NewClassifierService(dbPool, cfg)
	feedService := service.NewFeedService(dbPool)
	syncService := service.NewSyncService(dbPool, installmentService, classifierService, feedService)
	scheduler := jobs.NewScheduler(dbPool, syncService, userRepo, encryptionService, cfg)
	scheduler.Start()

	// 6. Registrar rotas
	router.Setup(e, dbPool, cfg)

	// Rota de Health Check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	// 7. Iniciar servidor com Graceful Shutdown
	go func() {
		port := cfg.Port
		if port == "" {
			port = "8080"
		}
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("Desligando o servidor... Erro: %v", err)
		}
	}()

	// Aguardar sinal de interrupção para desligar graciosamente
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Iniciando desligamento gracioso...")
	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	log.Println("Servidor finalizado com sucesso!")
}

package jobs

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

// Scheduler gerencia os jobs recorrentes do sistema.
type Scheduler struct {
	db                *pgxpool.Pool
	syncService       *service.SyncService
	userRepo          repository.UserRepository
	encryptionService *service.EncryptionService
	cron              *cron.Cron
	cfg               *config.Config
}

// NewScheduler cria uma nova instância de Scheduler.
func NewScheduler(db *pgxpool.Pool, syncService *service.SyncService, userRepo repository.UserRepository, encryptionService *service.EncryptionService, cfg *config.Config) *Scheduler {
	// Configura o cron para usar o fuso horário de Brasília (UTC-3)
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Printf("Aviso: erro ao carregar timezone America/Sao_Paulo, usando UTC: %v", err)
		loc = time.UTC
	}

	return &Scheduler{
		db:                db,
		syncService:       syncService,
		userRepo:          userRepo,
		encryptionService: encryptionService,
		cron:              cron.New(cron.WithLocation(loc)),
		cfg:               cfg,
	}
}

// Start inicia o agendador de tarefas.
func (s *Scheduler) Start() {
	// Agenda sincronização diária às 00:30 e dispara agentes
	_, err := s.cron.AddFunc("30 0 * * *", func() {
		log.Println("[Job] Iniciando sincronização diária de todos os usuários...")
		s.SyncAllUsers()
		log.Println("[Job] Disparando agentes de relatório...")
		s.TriggerAgents()
	})

	if err != nil {
		log.Fatalf("Erro ao agendar job de sincronização: %v", err)
	}

	// Agenda self-ping a cada 10 minutos para evitar sleep do Render
	_, err = s.cron.AddFunc("*/10 * * * *", func() {
		s.KeepAlive()
	})

	if err != nil {
		log.Printf("Aviso: erro ao agendar job de keepalive: %v", err)
	}

	s.cron.Start()
	log.Println("Scheduler iniciado com sucesso!")
}

// KeepAlive faz pings no próprio servidor e no serviço de agentes para evitar cold start.
func (s *Scheduler) KeepAlive() {
	client := &http.Client{Timeout: 5 * time.Second}

	// 1. Ping em si mesmo
	if s.cfg.SelfURL != "" {
		resp, err := client.Get(fmt.Sprintf("%s/health", s.cfg.SelfURL))
		if err != nil {
			log.Printf("[KeepAlive] Erro ao pingar self: %v", err)
		} else {
			resp.Body.Close()
		}
	}

	// 2. Ping no serviço de agentes
	if s.cfg.AgentsServiceURL != "" {
		resp, err := client.Get(fmt.Sprintf("%s/health", s.cfg.AgentsServiceURL))
		if err != nil {
			log.Printf("[KeepAlive] Erro ao pingar agentes: %v", err)
		} else {
			resp.Body.Close()
		}
	}
}

// TriggerAgents dispara os agentes de IA para todos os usuários ativos.
func (s *Scheduler) TriggerAgents() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	rows, err := s.db.Query(ctx, "SELECT DISTINCT user_id FROM connected_accounts")
	if err != nil {
		return
	}
	defer rows.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}

		// Dispara agente diário
		go func(id string) {
			resp, err := client.Post(fmt.Sprintf("%s/agents/daily/%s", s.cfg.AgentsServiceURL, id), "application/json", nil)
			if err == nil {
				resp.Body.Close()
			}
		}(userID)

		// Se for segunda-feira, dispara agente semanal
		if time.Now().Weekday() == time.Monday {
			go func(id string) {
				resp, err := client.Post(fmt.Sprintf("%s/agents/weekly/%s", s.cfg.AgentsServiceURL, id), "application/json", nil)
				if err == nil {
					resp.Body.Close()
				}
			}(userID)
		}

		// Se for dia 1 do mês, dispara agente mensal para fechar o mês anterior
		if time.Now().Day() == 1 {
			go func(id string) {
				resp, err := client.Post(fmt.Sprintf("%s/agents/monthly/%s", s.cfg.AgentsServiceURL, id), "application/json", nil)
				if err == nil {
					resp.Body.Close()
				}
			}(userID)
		}
	}
}

// Stop para o agendador.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// SyncAllUsers busca todos os usuários com contas conectadas e sincroniza.
func (s *Scheduler) SyncAllUsers() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	// Busca usuários que possuem pelo menos uma conta conectada
	query := `SELECT DISTINCT user_id FROM connected_accounts`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		log.Printf("[Job] Erro ao buscar usuários para sincronização: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}

		// Obter cliente pluggy para o usuário
		pluggyClient, err := s.getPluggyClientForUser(ctx, userID)
		if err != nil {
			log.Printf("[Job] Erro ao obter cliente pluggy para usuário %s: %v", userID, err)
			continue
		}

		// Para cada usuário, buscamos os item_ids únicos para sincronizar
		itemRows, err := s.db.Query(ctx, "SELECT DISTINCT pluggy_item_id FROM connected_accounts WHERE user_id = $1", userID)
		if err != nil {
			continue
		}

		for itemRows.Next() {
			var itemID string
			if err := itemRows.Scan(&itemID); err != nil {
				continue
			}

			log.Printf("[Job] Sincronizando item %s para usuário %s", itemID, userID)
			if _, err := s.syncService.SyncItem(ctx, userID, itemID, pluggyClient); err != nil {
				log.Printf("[Job] Erro ao sincronizar item %s: %v", itemID, err)
			}
		}
		itemRows.Close()
	}
}

func (s *Scheduler) getPluggyClientForUser(ctx context.Context, userID string) (*pluggy.Client, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado")
	}

	clientID := user.PluggyClientID
	var clientSecret string

	// Se o usuário tem chaves próprias
	if user.PluggyClientID != "" && user.PluggyClientSecretEncrypted != "" {
		decrypted, err := s.encryptionService.Decrypt(user.PluggyClientSecretEncrypted)
		if err == nil {
			clientSecret = decrypted
		}
	}

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("credenciais da Pluggy não configuradas para o usuário %s", userID)
	}

	return pluggy.NewClient(clientID, clientSecret), nil
}

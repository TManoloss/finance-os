package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type SyncHandler struct {
	db                *pgxpool.Pool
	syncService       *service.SyncService
	userRepo          repository.UserRepository
	encryptionService *service.EncryptionService
	cfg               *config.Config
}

func NewSyncHandler(
	db *pgxpool.Pool,
	syncService *service.SyncService,
	userRepo repository.UserRepository,
	encryptionService *service.EncryptionService,
	cfg *config.Config,
) *SyncHandler {
	return &SyncHandler{
		db:                db,
		syncService:       syncService,
		userRepo:          userRepo,
		encryptionService: encryptionService,
		cfg:               cfg,
	}
}

// SyncAll dispara a sincronização para todos os usuários com contas conectadas.
func (h *SyncHandler) SyncAll(c echo.Context) error {
	// 1. Validar Sync Secret
	secret := c.Request().Header.Get("X-Sync-Secret")
	if secret == "" || secret != h.cfg.SyncSecret {
		return response.Error(c, http.StatusUnauthorized, "sync secret inválido")
	}

	start := time.Now()
	log.Printf("[SyncInternal] Iniciando sincronização global disparada por %s", c.RealIP())

	// 2. Buscar usuários que possuem contas conectadas
	query := `SELECT DISTINCT user_id FROM connected_accounts`
	rows, err := h.db.Query(c.Request().Context(), query)
	if err != nil {
		log.Printf("[SyncInternal] Erro ao buscar usuários: %v", err)
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar usuários para sync")
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err == nil {
			userIDs = append(userIDs, uid)
		}
	}

	// 3. Sync em paralelo com semáforo (máx 3 goroutines)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)
	
	results := struct {
		SyncedUsers      int `json:"synced_users"`
		TotalImported    int `json:"total_transactions_imported"`
		ErrorsCount      int `json:"errors_count"`
		Errors           []map[string]string `json:"errors"`
		DurationMS       int64 `json:"duration_ms"`
		TriggeredAt      string `json:"triggered_at"`
	}{
		SyncedUsers: len(userIDs),
		TriggeredAt: start.Format(time.RFC3339),
	}

	var resultsMu sync.Mutex

	for _, uid := range userIDs {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			sem <- struct{}{}        // adquire
			defer func() { <-sem }() // libera

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// Obter cliente Pluggy para o usuário
			pluggyClient, err := h.getPluggyClientForUser(ctx, userID)
			if err != nil {
				resultsMu.Lock()
				results.ErrorsCount++
				results.Errors = append(results.Errors, map[string]string{
					"user_id": userID,
					"error":   fmt.Sprintf("erro ao obter cliente pluggy: %v", err),
				})
				resultsMu.Unlock()
				return
			}

			// Sincronizar
			count, err := h.syncService.SyncUserAccounts(ctx, userID, pluggyClient)
			
			resultsMu.Lock()
			if err != nil {
				results.ErrorsCount++
				results.Errors = append(results.Errors, map[string]string{
					"user_id": userID,
					"error":   err.Error(),
				})
			}
			results.TotalImported += count
			resultsMu.Unlock()
		}(uid)
	}

	wg.Wait()
	results.DurationMS = time.Since(start).Milliseconds()

	// 4. Salvar log do sync
	_, err = h.db.Exec(context.Background(), `
		INSERT INTO sync_logs (triggered_by, synced_users, transactions_imported, errors_count, errors_detail, duration_ms, started_at, finished_at)
		VALUES ('cron', $1, $2, $3, $4, $5, $6, NOW())`,
		results.SyncedUsers, results.TotalImported, results.ErrorsCount, results.Errors, results.DurationMS, start)
	if err != nil {
		log.Printf("[SyncInternal] Erro ao salvar sync_log: %v", err)
	}

	log.Printf("[SyncInternal] Sincronização global finalizada em %dms. Usuários: %d, Transações: %d, Erros: %d", 
		results.DurationMS, results.SyncedUsers, results.TotalImported, results.ErrorsCount)

	return response.Success(c, http.StatusOK, results)
}

// Status retorna o status da última sincronização.
func (h *SyncHandler) Status(c echo.Context) error {
	secret := c.Request().Header.Get("X-Sync-Secret")
	if secret == "" || secret != h.cfg.SyncSecret {
		return response.Error(c, http.StatusUnauthorized, "sync secret inválido")
	}

	var res struct {
		LastSyncAt           time.Time `json:"last_sync_at"`
		LastSyncDurationMS   int       `json:"last_sync_duration_ms"`
		LastSyncUsers        int       `json:"last_sync_users"`
		LastSyncTransactions int       `json:"last_sync_transactions"`
		NextScheduledSync    string    `json:"next_scheduled_sync"`
	}

	query := `
		SELECT started_at, duration_ms, synced_users, transactions_imported 
		FROM sync_logs 
		WHERE triggered_by = 'cron'
		ORDER BY started_at DESC LIMIT 1`
	
	err := h.db.QueryRow(c.Request().Context(), query).Scan(
		&res.LastSyncAt, &res.LastSyncDurationMS, &res.LastSyncUsers, &res.LastSyncTransactions,
	)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "nenhum log de sincronização encontrado")
	}

	// Estimativa simples para o próximo sync (assumindo horários fixos do cron-job.org)
	now := time.Now()
	res.NextScheduledSync = "Próximo horário agendado (07:00, 13:00, 19:00 ou 23:30)"

	return response.Success(c, http.StatusOK, res)
}

func (h *SyncHandler) getPluggyClientForUser(ctx context.Context, userID string) (*pluggy.Client, error) {
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado")
	}

	clientID := user.PluggyClientID
	var clientSecret string

	if user.PluggyClientID != "" && user.PluggyClientSecretEncrypted != "" {
		decrypted, err := h.encryptionService.Decrypt(user.PluggyClientSecretEncrypted)
		if err == nil {
			clientSecret = decrypted
		}
	}

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("credenciais da Pluggy não configuradas")
	}

	return pluggy.NewClient(clientID, clientSecret), nil
}

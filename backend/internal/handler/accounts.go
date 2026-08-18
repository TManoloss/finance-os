package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type AccountsHandler struct {
	db                *pgxpool.Pool
	syncService       *service.SyncService
	encryptionService *service.EncryptionService
	userRepo          repository.UserRepository
	cfg               *config.Config
}

func NewAccountsHandler(
	db *pgxpool.Pool,
	syncService *service.SyncService,
	encryptionService *service.EncryptionService,
	userRepo repository.UserRepository,
	cfg *config.Config,
) *AccountsHandler {
	return &AccountsHandler{
		db:                db,
		syncService:       syncService,
		encryptionService: encryptionService,
		userRepo:          userRepo,
		cfg:               cfg,
	}
}

// SavePluggyKeys salva as credenciais da Pluggy para o usuário.
// Antes de persistir, valida que as credenciais são aceitas pela API da Pluggy.
func (h *AccountsHandler) SavePluggyKeys(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	if req.ClientID == "" || req.ClientSecret == "" {
		return response.Error(c, http.StatusBadRequest, "client_id e client_secret são obrigatórios")
	}

	// Validar credenciais na Pluggy antes de salvar
	// Etapa 1: Testar autenticação (POST /auth)
	testClient := pluggy.NewClient(req.ClientID, req.ClientSecret)
	if err := testClient.Authenticate(); err != nil {
		log.Printf("[SavePluggyKeys] Credenciais inválidas para user %s: %v", userID, err)
		return response.Error(c, http.StatusBadRequest, "Credenciais inválidas. A Pluggy rejeitou o Client ID e Client Secret informados. Verifique se copiou corretamente no painel da Pluggy.")
	}

	// Etapa 2: Testar operação real (POST /connect_token)
	// Algumas credenciais passam no auth mas não funcionam na prática
	if _, err := testClient.CreateConnectToken(nil); err != nil {
		log.Printf("[SavePluggyKeys] Credenciais autenticam mas não operam para user %s: %v", userID, err)
		return response.Error(c, http.StatusBadRequest, "Suas credenciais autenticaram, mas a Pluggy não permitiu operações. Verifique se sua conta na Pluggy está ativa e se você está usando as credenciais do ambiente correto (Produção vs Sandbox).")
	}
	log.Printf("[SavePluggyKeys] Credenciais válidas e operacionais para user %s, salvando...", userID)

	encryptedSecret, err := h.encryptionService.Encrypt(req.ClientSecret)
	if err != nil {
		log.Printf("Erro ao criptografar secret: %v", err)
		return response.Error(c, http.StatusInternalServerError, "erro ao processar credenciais")
	}

	err = h.userRepo.UpdatePluggyCredentials(c.Request().Context(), userID, req.ClientID, encryptedSecret)
	if err != nil {
		log.Printf("Erro ao salvar credenciais no banco: %v", err)
		return response.Error(c, http.StatusInternalServerError, "erro ao salvar credenciais")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"message": "credenciais validadas e salvas com sucesso",
	})
}

// SaveLLMKeys salva as credenciais de IA (Groq e Gemini) para o usuário.
func (h *AccountsHandler) SaveLLMKeys(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		GroqAPIKey   string `json:"groq_api_key"`
		GeminiAPIKey string `json:"gemini_api_key"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	var groqEncrypted, geminiEncrypted string
	var err error

	if req.GroqAPIKey != "" {
		groqEncrypted, err = h.encryptionService.Encrypt(req.GroqAPIKey)
		if err != nil {
			log.Printf("Erro ao criptografar chave do Groq: %v", err)
			return response.Error(c, http.StatusInternalServerError, "erro ao processar credenciais")
		}
	}

	if req.GeminiAPIKey != "" {
		geminiEncrypted, err = h.encryptionService.Encrypt(req.GeminiAPIKey)
		if err != nil {
			log.Printf("Erro ao criptografar chave do Gemini: %v", err)
			return response.Error(c, http.StatusInternalServerError, "erro ao processar credenciais")
		}
	}

	err = h.userRepo.UpdateLLMCredentials(c.Request().Context(), userID, groqEncrypted, geminiEncrypted)
	if err != nil {
		log.Printf("Erro ao salvar credenciais de IA no banco: %v", err)
		return response.Error(c, http.StatusInternalServerError, "erro ao salvar credenciais")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"message": "credenciais de IA salvas com sucesso",
	})
}

// ListAccounts lista as contas conectadas do usuário.
func (h *AccountsHandler) ListAccounts(c echo.Context) error {
	userID := c.Get("user_id").(string)
	log.Printf("[Accounts] Listando contas para o usuário: %s", userID)

	query := `
		SELECT 
			id, 
			institution_name, 
			institution_logo, 
			institution_color, 
			account_type, 
			balance, 
			currency, 
			last_synced_at, 
			pluggy_item_id,
			account_name,
			account_number_last4,
			connection_label,
			subtype,
			close_day, 
			due_day 
		FROM connected_accounts 
		WHERE user_id = $1 
		ORDER BY institution_name ASC`

	rows, err := h.db.Query(c.Request().Context(), query, userID)
	if err != nil {
		log.Printf("[Accounts] Erro ao executar query para user %s: %v", userID, err)
		return response.Error(c, http.StatusInternalServerError, "erro interno ao buscar contas")
	}
	defer rows.Close()

	var accounts []map[string]interface{}
	for rows.Next() {
		var (
			id, instName, accType, currency string
			instLogo, instColor, pluggyID   *string
			accountName, accountLast4       *string
			connectionLabel, subtype        *string
			balance                         float64
			lastSynced                      *time.Time
			closeDay, dueDay                *int
		)

		err := rows.Scan(
			&id, &instName, &instLogo, &instColor,
			&accType, &balance, &currency, &lastSynced,
			&pluggyID, &accountName, &accountLast4, &connectionLabel, &subtype,
			&closeDay, &dueDay,
		)
		if err != nil {
			log.Printf("[Accounts] Erro ao escanear linha para user %s: %v", userID, err)
			continue
		}

		// Garante valores padrão para campos obrigatórios na UI
		if instName == "" {
			instName = "Instituição Desconhecida"
		}
		if accType == "" {
			accType = "CHECKING"
		}
		if currency == "" {
			currency = "BRL"
		}

		accounts = append(accounts, map[string]interface{}{
			"id":                id,
			"institution_name":  instName,
			"institution_logo":  instLogo,
			"institution_color": instColor,
			"account_type":      accType,
			"balance":           balance,
			"currency":          currency,
			"last_synced_at":    lastSynced,
			"pluggy_item_id":    pluggyID,
			"account_name":      accountName,
			"account_last4":     accountLast4,
			"connection_label":  connectionLabel,
			"subtype":           subtype,
			"close_day":         closeDay,
			"due_day":           dueDay,
		})
	}

	return response.Success(c, http.StatusOK, accounts)
}

// CreateManual cria uma conta ou cartão sem integração Open Finance.
func (h *AccountsHandler) CreateManual(c echo.Context) error {
	userID := c.Get("user_id").(string)
	var req struct {
		Name, Type       string
		Balance          float64 `json:"balance"`
		CloseDay, DueDay int     `json:"close_day"`
	}
	if err := c.Bind(&req); err != nil || req.Name == "" || (req.Type != "CHECKING" && req.Type != "SAVINGS" && req.Type != "CREDIT") {
		return response.Error(c, http.StatusBadRequest, "nome e tipo de conta válidos são obrigatórios")
	}
	var id string
	err := h.db.QueryRow(c.Request().Context(), `INSERT INTO connected_accounts (user_id, institution_name, account_type, balance, close_day, due_day) VALUES ($1,$2,$3,$4,NULLIF($5,0),NULLIF($6,0)) RETURNING id`, userID, req.Name, req.Type, req.Balance, req.CloseDay, req.DueDay).Scan(&id)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao criar conta")
	}
	return response.Success(c, http.StatusCreated, map[string]string{"id": id})
}

// UpdateAccountSettings atualiza configurações específicas da conta (ex: dia de fechamento do cartão).
func (h *AccountsHandler) UpdateAccountSettings(c echo.Context) error {
	userID := c.Get("user_id").(string)
	accountID := c.Param("id")

	var req struct {
		CloseDay int `json:"close_day"`
		DueDay   int `json:"due_day"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	query := `UPDATE connected_accounts SET close_day = $1, due_day = $2 WHERE id = $3 AND user_id = $4`
	_, err := h.db.Exec(c.Request().Context(), query, req.CloseDay, req.DueDay, accountID, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao atualizar configurações da conta")
	}

	return response.Success(c, http.StatusOK, map[string]string{"message": "configurações atualizadas"})
}

// ConnectToken gera um token para o widget de conexão do Pluggy.
func (h *AccountsHandler) ConnectToken(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		ItemID string `json:"item_id"`
	}
	_ = c.Bind(&req) // Opcional

	pluggyClient, err := h.getPluggyClientForUser(c.Request().Context(), userID)
	if err != nil {
		log.Printf("[ConnectToken] Falha ao obter client Pluggy para user %s: %v", userID, err)
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	token, err := pluggyClient.CreateConnectToken(&req.ItemID)
	if err != nil {
		log.Printf("[ConnectToken] Falha ao gerar connect token para user %s: %v", userID, err)
		return response.Error(c, http.StatusBadGateway, err.Error())
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"accessToken": token,
	})
}

// Sync dispara a sincronização manual das contas.
func (h *AccountsHandler) Sync(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		ItemID string `json:"item_id"`
		Force  *bool  `json:"force"`
	}
	_ = c.Bind(&req)

	if req.ItemID == "" {
		// Rate limit check: max 1 manual sync per user every 30 minutes
		var lastSync time.Time
		err := h.db.QueryRow(c.Request().Context(), `
			SELECT MAX(last_synced_at) FROM connected_accounts WHERE user_id = $1
		`, userID).Scan(&lastSync)

		if err == nil && !lastSync.IsZero() && time.Since(lastSync) < 30*time.Minute {
			return response.Error(c, http.StatusTooManyRequests, "sincronização manual permitida apenas a cada 30 minutos")
		}
	}

	pluggyClient, err := h.getPluggyClientForUser(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	var runID string
	if err := h.db.QueryRow(c.Request().Context(), `
		INSERT INTO sync_logs (user_id, item_id, triggered_by, status, started_at)
		VALUES ($1, NULLIF($2, ''), 'manual', 'running', NOW())
		RETURNING id
	`, userID, req.ItemID).Scan(&runID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "não foi possível registrar a sincronização")
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		start := time.Now()
		var saved int
		var err error

		if req.ItemID != "" {
			saved, err = h.syncService.SyncItem(ctx, userID, req.ItemID, pluggyClient, syncForce(req.Force))
		} else {
			saved, err = h.syncService.SyncUserAccounts(ctx, userID, pluggyClient, true)
		}

		duration := time.Since(start).Milliseconds()

		// Log na tabela sync_logs
		errorsCount := 0
		var errorsDetail string
		if err != nil {
			errorsCount = 1

			// Usa json.Marshal para escapar corretamente as aspas e quebras de linha
			errObj := []map[string]string{
				{"user_id": userID, "error": err.Error()},
			}
			jsonBytes, _ := json.Marshal(errObj)
			errorsDetail = string(jsonBytes)
		} else {
			errorsDetail = "[]"
		}

		_, dbErr := h.db.Exec(context.Background(), `
			UPDATE sync_logs SET synced_users = 1, transactions_imported = $1,
				errors_count = $2, errors_detail = $3::jsonb, duration_ms = $4,
				status = $5, finished_at = NOW()
			WHERE id = $6 AND user_id = $7
		`, saved, errorsCount, errorsDetail, duration, syncRunStatus(err), runID, userID)

		if dbErr != nil {
			log.Printf("Erro ao salvar log de sync_logs: %v", dbErr)
		}

		if err != nil {
			log.Printf("Erro na sincronização assíncrona para user %s: %v", userID, err)
		}
	}()

	return response.Success(c, http.StatusAccepted, map[string]interface{}{
		"message": "sincronização iniciada com sucesso",
		"run_id":  runID,
		"status":  "running",
	})
}

func syncForce(force *bool) bool {
	return force == nil || *force
}

func syncRunStatus(err error) string {
	if err == nil {
		return "completed"
	}
	if errors.Is(err, service.ErrPartialSync) {
		return "partial"
	}
	return "failed"
}

// SyncStatus retorna uma execução manual pertencente ao Usuário autenticado.
func (h *AccountsHandler) SyncStatus(c echo.Context) error {
	userID := c.Get("user_id").(string)
	runID := c.Param("run_id")
	var itemID, errorMessage *string
	var status string
	var transactions, errorsCount int
	var startedAt time.Time
	var finishedAt *time.Time
	err := h.db.QueryRow(c.Request().Context(), `
		SELECT item_id, status, transactions_imported, errors_count,
		       errors_detail->0->>'error', started_at, finished_at
		FROM sync_logs WHERE id = $1 AND user_id = $2 AND triggered_by = 'manual'
	`, runID, userID).Scan(
		&itemID, &status, &transactions, &errorsCount,
		&errorMessage, &startedAt, &finishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return response.Error(c, http.StatusNotFound, "sincronização não encontrada")
	}
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao consultar sincronização")
	}
	return response.Success(c, http.StatusOK, map[string]interface{}{
		"id":                    runID,
		"item_id":               itemID,
		"status":                status,
		"transactions_imported": transactions,
		"errors_count":          errorsCount,
		"error":                 errorMessage,
		"started_at":            startedAt,
		"finished_at":           finishedAt,
	})
}

// ImportSummary retorna o comprovante real da importação de uma conexão Pluggy.
func (h *AccountsHandler) ImportSummary(c echo.Context) error {
	userID := c.Get("user_id").(string)
	itemID := c.QueryParam("item_id")
	if itemID == "" {
		return response.Error(c, http.StatusBadRequest, "item_id é obrigatório")
	}

	var accounts, transactions, needsReview int
	var fromDate, toDate *time.Time
	var lastSynced *time.Time
	err := h.db.QueryRow(c.Request().Context(), `
		SELECT COUNT(DISTINCT a.id), COUNT(t.id), MIN(t.date), MAX(t.date),
		       COUNT(t.id) FILTER (WHERE t.needs_review), MAX(a.last_synced_at)
		FROM connected_accounts a
		LEFT JOIN transactions t ON t.account_id = a.id
		WHERE a.user_id = $1 AND a.pluggy_item_id = $2
	`, userID, itemID).Scan(
		&accounts, &transactions, &fromDate, &toDate, &needsReview, &lastSynced,
	)
	if err != nil {
		log.Printf("[ImportSummary] Erro ao montar comprovante para user %s: %v", userID, err)
		return response.Error(c, http.StatusInternalServerError, "erro ao consultar a importação")
	}

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"status":                importStatus(lastSynced),
		"accounts_found":        accounts,
		"transactions_imported": transactions,
		"period_from":           fromDate,
		"period_to":             toDate,
		"needs_review":          needsReview,
		"updated_at":            lastSynced,
	})
}

func importStatus(lastSynced *time.Time) string {
	if lastSynced == nil {
		return "processing"
	}
	return "completed"
}

// DeleteAccount desconecta e remove uma conta e todos os seus dados associados.
func (h *AccountsHandler) DeleteAccount(c echo.Context) error {
	userID := c.Get("user_id").(string)
	accountID := c.Param("id")

	// 1. Verificar se a conta pertence ao usuário
	var exists bool
	err := h.db.QueryRow(c.Request().Context(), "SELECT EXISTS(SELECT 1 FROM connected_accounts WHERE id = $1 AND user_id = $2)", accountID, userID).Scan(&exists)
	if err != nil || !exists {
		return response.Error(c, http.StatusNotFound, "conta não encontrada ou permissão negada")
	}

	// 2. Deletar a conta (O banco de dados lidará com o delete cascade para transações e parcelas)
	_, err = h.db.Exec(c.Request().Context(), "DELETE FROM connected_accounts WHERE id = $1", accountID)
	if err != nil {
		log.Printf("Erro ao deletar conta %s: %v", accountID, err)
		return response.Error(c, http.StatusInternalServerError, "erro ao desconectar conta")
	}

	log.Printf("[Accounts] Conta %s removida pelo usuário %s", accountID, userID)
	return response.Success(c, http.StatusOK, map[string]string{
		"message": "conta desconectada com sucesso",
	})
}

// DeleteConnection revoga uma conexão Pluggy e remove apenas as Contas importadas por ela.
func (h *AccountsHandler) DeleteConnection(c echo.Context) error {
	userID := c.Get("user_id").(string)
	itemID := c.Param("item_id")
	var exists bool
	if err := h.db.QueryRow(c.Request().Context(), `
		SELECT EXISTS(
			SELECT 1 FROM connected_accounts
			WHERE user_id = $1 AND pluggy_item_id = $2
		)
	`, userID, itemID).Scan(&exists); err != nil || !exists {
		return response.Error(c, http.StatusNotFound, "conexão não encontrada")
	}

	client, err := h.getPluggyClientForUser(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
	if err := client.DeleteItem(itemID); err != nil {
		log.Printf("[Accounts] Falha ao revogar conexão %s do user %s: %v", itemID, userID, err)
		return response.Error(c, http.StatusBadGateway, "não foi possível revogar a conexão na Pluggy")
	}
	if _, err := h.db.Exec(c.Request().Context(), `
		DELETE FROM connected_accounts WHERE user_id = $1 AND pluggy_item_id = $2
	`, userID, itemID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "conexão revogada, mas os dados locais não puderam ser removidos")
	}
	return response.Success(c, http.StatusOK, map[string]string{
		"message": "conexão removida com sucesso",
	})
}

// UpdateConnectionLabel permite diferenciar Conexões da mesma instituição, como PF e PJ.
func (h *AccountsHandler) UpdateConnectionLabel(c echo.Context) error {
	userID := c.Get("user_id").(string)
	itemID := c.Param("item_id")
	var req struct {
		Label string `json:"label"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}
	req.Label = strings.TrimSpace(req.Label)
	if len([]rune(req.Label)) > 50 {
		return response.Error(c, http.StatusBadRequest, "o nome da conexão deve ter no máximo 50 caracteres")
	}
	result, err := h.db.Exec(c.Request().Context(), `
		UPDATE connected_accounts SET connection_label = NULLIF($1, '')
		WHERE user_id = $2 AND pluggy_item_id = $3
	`, req.Label, userID, itemID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao nomear conexão")
	}
	if result.RowsAffected() == 0 {
		return response.Error(c, http.StatusNotFound, "conexão não encontrada")
	}
	return response.Success(c, http.StatusOK, map[string]string{"label": req.Label})
}

func (h *AccountsHandler) getPluggyClientForUser(ctx context.Context, userID string) (*pluggy.Client, error) {
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado")
	}

	clientID := user.PluggyClientID
	var clientSecret string

	log.Printf("[DEBUG] Verificando chaves para user %s: has_id=%v", userID, clientID != "")

	// Se o usuário tem chaves próprias
	if user.PluggyClientID != "" && user.PluggyClientSecretEncrypted != "" {
		decrypted, err := h.encryptionService.Decrypt(user.PluggyClientSecretEncrypted)
		if err == nil {
			clientSecret = decrypted
		} else {
			log.Printf("[DEBUG] Falha ao descriptografar secret para user %s: %v", userID, err)
		}
	} else if user.PluggyClientID != "" {
		log.Printf("[DEBUG] User %s tem ClientID mas SecretEncrypted está vazio", userID)
	}

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("credenciais da Pluggy não configuradas. Por favor, configure suas chaves nas configurações")
	}

	return pluggy.NewClient(clientID, clientSecret), nil
}

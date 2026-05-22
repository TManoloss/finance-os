package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/pluggy"
	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
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
			instLogo, instColor, pluggyID *string
			balance float64
			lastSynced *time.Time
			closeDay, dueDay *int
		)

		err := rows.Scan(
			&id, &instName, &instLogo, &instColor, 
			&accType, &balance, &currency, &lastSynced, 
			&pluggyID, &closeDay, &dueDay,
		)
		if err != nil {
			log.Printf("[Accounts] Erro ao escanear linha para user %s: %v", userID, err)
			continue
		}

		// Garante valores padrão para campos obrigatórios na UI
		if instName == "" { instName = "Instituição Desconhecida" }
		if accType == "" { accType = "CHECKING" }
		if currency == "" { currency = "BRL" }

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
			"close_day":         closeDay,
			"due_day":           dueDay,
		})
	}

	return response.Success(c, http.StatusOK, accounts)
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
		return response.Error(c, http.StatusBadGateway, "Falha ao conectar com a Pluggy. Suas credenciais podem estar expiradas ou inválidas. Atualize-as nas Configurações.")
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

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		start := time.Now()
		var saved int
		var err error

		if req.ItemID != "" {
			saved, err = h.syncService.SyncItem(ctx, userID, req.ItemID, pluggyClient)
		} else {
			saved, err = h.syncService.SyncUserAccounts(ctx, userID, pluggyClient)
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
		
		h.db.Exec(context.Background(), `
			INSERT INTO sync_logs (triggered_by, synced_users, transactions_imported, errors_count, errors_detail, duration_ms, started_at, finished_at)
			VALUES ('manual', 1, $1, $2, $3::jsonb, $4, $5, NOW())
		`, saved, errorsCount, errorsDetail, duration, start)

		if err != nil {
			log.Printf("Erro na sincronização assíncrona para user %s: %v", userID, err)
		}
	}()

	return response.Success(c, http.StatusAccepted, map[string]interface{}{
		"message": "sincronização iniciada com sucesso",
	})
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

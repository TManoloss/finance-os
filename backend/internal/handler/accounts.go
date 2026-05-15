package handler

import (
	"context"
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
		"message": "credenciais salvas com sucesso",
	})
}

// ListAccounts lista as contas conectadas do usuário.
func (h *AccountsHandler) ListAccounts(c echo.Context) error {
	userID := c.Get("user_id").(string)
	query := `
		SELECT id, institution_name, institution_logo, institution_color, account_type, balance, currency, last_synced_at, pluggy_item_id, close_day, due_day 
		FROM connected_accounts 
		WHERE user_id = $1 
		ORDER BY institution_name ASC`
	rows, err := h.db.Query(c.Request().Context(), query, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar contas")
	}
	defer rows.Close()

	var accounts []map[string]interface{}
	for rows.Next() {
		var acc struct {
			ID               string
			InstitutionName  string
			InstitutionLogo  *string
			InstitutionColor *string
			AccountType      string
			Balance          float64
			Currency         string
			LastSyncedAt     *time.Time
			PluggyItemID     *string
			CloseDay         int
			DueDay           int
		}
		err := rows.Scan(&acc.ID, &acc.InstitutionName, &acc.InstitutionLogo, &acc.InstitutionColor, &acc.AccountType, &acc.Balance, &acc.Currency, &acc.LastSyncedAt, &acc.PluggyItemID, &acc.CloseDay, &acc.DueDay)
		if err != nil {
			log.Printf("Erro ao escanear conta: %v", err)
			continue
		}
		
		accounts = append(accounts, map[string]interface{}{
			"id":                acc.ID,
			"institution_name":  acc.InstitutionName,
			"institution_logo":  acc.InstitutionLogo,
			"institution_color": acc.InstitutionColor,
			"account_type":      acc.AccountType,
			"balance":           acc.Balance,
			"currency":          acc.Currency,
			"last_synced_at":    acc.LastSyncedAt,
			"pluggy_item_id":    acc.PluggyItemID,
			"close_day":         acc.CloseDay,
			"due_day":           acc.DueDay,
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
		return response.Error(c, http.StatusUnauthorized, err.Error())
	}

	token, err := pluggyClient.CreateConnectToken(&req.ItemID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao gerar connect token")
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
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	if req.ItemID == "" {
		return response.Error(c, http.StatusBadRequest, "o item_id é obrigatório")
	}

	pluggyClient, err := h.getPluggyClientForUser(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, err.Error())
	}

	// Executa em background para não travar a request
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := h.syncService.SyncItem(ctx, userID, req.ItemID, pluggyClient); err != nil {
			log.Printf("Erro na sincronização assíncrona para user %s: %v", userID, err)
		}
	}()

	return response.Success(c, http.StatusAccepted, map[string]string{
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

	// Se o usuário tem chaves próprias
	if user.PluggyClientID != "" && user.PluggyClientSecretEncrypted != "" {
		decrypted, err := h.encryptionService.Decrypt(user.PluggyClientSecretEncrypted)
		if err == nil {
			clientSecret = decrypted
		}
	}

	// Se ainda não temos chaves (ou falhou a descriptografia), tenta o fallback global
	if clientID == "" || clientSecret == "" {
		if h.cfg.PluggyClientID != "" && h.cfg.PluggyClientSecret != "" {
			clientID = h.cfg.PluggyClientID
			clientSecret = h.cfg.PluggyClientSecret
			log.Printf("[Accounts] Usando chaves globais de fallback para user %s", userID)
		}
	}

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("credenciais da Pluggy não configuradas")
	}

	return pluggy.NewClient(clientID, clientSecret), nil
}

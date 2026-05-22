package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/finance-os/backend/internal/pluggy"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncService coordena a sincronização de dados da Pluggy para o banco local.
type SyncService struct {
	db                 *pgxpool.Pool
	installmentService *InstallmentsService
	classifierService  *ClassifierService
	feedService        *FeedService
}

// NewSyncService cria uma nova instância de SyncService.
func NewSyncService(db *pgxpool.Pool, installmentService *InstallmentsService, classifierService *ClassifierService, feedService *FeedService) *SyncService {
	return &SyncService{
		db:                 db,
		installmentService: installmentService,
		classifierService:  classifierService,
		feedService:        feedService,
	}
}

// SyncItem realiza a sincronização completa de um Item (contas e transações).
// Quando chamado logo após a conexão via widget, o item pode estar em status
// UPDATING por até 2 minutos. Fazemos polling até ficar UPDATED.
func (s *SyncService) SyncItem(ctx context.Context, userID string, itemID string, pluggyClient *pluggy.Client) (int, error) {
	log.Printf("[SyncItem] Iniciando sync do item %s para user %s", itemID, userID)

	// 1. Verificar o status do item na Pluggy
	// Para syncs de rotina, o item já existe e geralmente estará em status UPDATED.
	// Para novas conexões via widget, pode estar UPDATING por até 2 minutos.
	var item *pluggy.Item
	var err error
	maxAttempts := 18 // 18 x 10s = 3 minutos de timeout
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		item, err = pluggyClient.GetItem(itemID)
		if err != nil {
			log.Printf("[SyncItem] Tentativa %d/%d: erro ao buscar item %s: %v", attempt, maxAttempts, itemID, err)
			
			// Se o item não for encontrado (404), não adianta tentar novamente
			if strings.Contains(err.Error(), "status 404") {
				return 0, fmt.Errorf("item %s não foi encontrado na Pluggy (possivelmente desconectado).", itemID)
			}
			
			if attempt == maxAttempts {
				return 0, fmt.Errorf("item %s não encontrado na Pluggy após %d tentativas: %w", itemID, maxAttempts, err)
			}
			time.Sleep(10 * time.Second)
			continue
		}

		log.Printf("[SyncItem] Tentativa %d/%d: item %s status=%s", attempt, maxAttempts, itemID, item.Status)

		if item.Status == "UPDATED" || item.Status == "FINISHED" {
			break // Item pronto
		}

		if item.Status == "LOGIN_ERROR" || item.Status == "OUTDATED" || item.Status == "WAITING_USER_INPUT" {
			return 0, fmt.Errorf("item %s com erro de conexão (status: %s). O banco pode ter rejeitado a autenticação", itemID, item.Status)
		}

		// Status UPDATING, MERGING, etc
		// Se for a primeira tentativa e o item está atualizando (sync automático),
		// aguardamos até 15s antes de usar os dados existentes mesmo assim.
		// Isso evita timeout de 3min em syncs de rotina.
		if attempt >= 2 {
			log.Printf("[SyncItem] Item %s em status %s após %ds de espera. Prosseguindo com dados disponíveis.", itemID, item.Status, attempt*10)
			break
		}

		if attempt == maxAttempts {
			return 0, fmt.Errorf("item %s não ficou pronto após 3 minutos (status: %s)", itemID, item.Status)
		}
		time.Sleep(10 * time.Second)
	}

	// 2. Buscar detalhes do Conector (Logo e Cor)
	var logo, color string
	if item != nil {
		connector, err := pluggyClient.GetConnector(item.ConnectorID)
		if err == nil {
			logo = connector.ImageUrl
			color = connector.PrimaryColor
			log.Printf("[SyncItem] Conector: %s (logo=%s)", connector.Name, logo)
		} else {
			log.Printf("[SyncItem] Aviso: não foi possível buscar conector %d: %v", item.ConnectorID, err)
		}
	}

	// 3. Buscar contas do Item na Pluggy
	pluggyAccounts, err := pluggyClient.GetAccounts(itemID)
	if err != nil {
		return 0, fmt.Errorf("erro ao buscar contas na pluggy para item %s: %w", itemID, err)
	}

	log.Printf("[SyncItem] Item %s: encontradas %d contas", itemID, len(pluggyAccounts))

	if len(pluggyAccounts) == 0 {
		return 0, fmt.Errorf("item %s conectado mas retornou 0 contas. Pode ser que o banco não tenha informações disponíveis", itemID)
	}

	totalSaved := 0

	for _, pa := range pluggyAccounts {
		log.Printf("[SyncItem] Processando conta %s (%s/%s) - saldo: %.2f", pa.ID, pa.Name, pa.Type, pa.Balance)

		// 4. Upsert da conta no nosso banco com logo e cor
		accountID, err := s.upsertAccount(ctx, userID, pa, logo, color)
		if err != nil {
			log.Printf("[SyncItem] Erro ao sincronizar conta %s: %v", pa.ID, err)
			continue
		}

		// 5. Buscar transações dos últimos 90 dias
		// Adicionamos 1 dia ao 'to' para garantir que transações de hoje não sejam cortadas por fuso horário.
		to := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		from := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
		
		transactions, err := pluggyClient.GetTransactions(pa.ID, from, to)
		if err != nil {
			log.Printf("[SyncItem] Erro ao buscar transações da conta %s: %v", pa.ID, err)
			continue
		}

		log.Printf("[SyncItem] Conta %s: %d transações encontradas (período %s a %s)", pa.ID, len(transactions), from, to)

		// 6. Salvar transações (passando userID para classificação)
		savedTxs, err := s.saveTransactionsAndReturn(ctx, userID, accountID, transactions)
		if err != nil {
			log.Printf("[SyncItem] Erro ao salvar transações da conta %s: %v", pa.ID, err)
		}

		totalSaved += len(savedTxs)
		log.Printf("[SyncItem] Conta %s: %d transações salvas (novas)", pa.ID, len(savedTxs))

		// 7. Detectar parcelamentos
		if s.installmentService != nil && len(savedTxs) > 0 {
			if err := s.installmentService.ProcessTransactions(ctx, accountID, savedTxs); err != nil {
				log.Printf("[SyncItem] Erro ao processar parcelamentos: %v", err)
			}
		}

		// 8. Gerar eventos no feed
		if s.feedService != nil && len(savedTxs) > 0 {
			if err := s.feedService.GenerateEvents(ctx, userID, savedTxs); err != nil {
				log.Printf("[SyncItem] Erro ao gerar eventos no feed: %v", err)
			}
		}
		
		// 9. Atualizar last_synced_at
		s.db.Exec(ctx, "UPDATE connected_accounts SET last_synced_at = NOW() WHERE id = $1", accountID)
	}

	log.Printf("[SyncItem] Sync completo: item %s, user %s, total %d transações salvas", itemID, userID, totalSaved)
	return totalSaved, nil
}

// SyncUserAccounts realiza a sincronização de todas as contas conectadas de um usuário.
func (s *SyncService) SyncUserAccounts(ctx context.Context, userID string, pluggyClient *pluggy.Client) (int, error) {
	rows, err := s.db.Query(ctx, "SELECT DISTINCT pluggy_item_id FROM connected_accounts WHERE user_id = $1 AND pluggy_item_id IS NOT NULL", userID)
	if err != nil {
		return 0, fmt.Errorf("erro ao buscar items do usuário: %w", err)
	}
	defer rows.Close()

	var itemIDs []string
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err == nil {
			itemIDs = append(itemIDs, itemID)
		}
	}

	totalSaved := 0
	for _, itemID := range itemIDs {
		saved, err := s.SyncItem(ctx, userID, itemID, pluggyClient)
		if err != nil {
			log.Printf("Erro ao sincronizar item %s do usuário %s: %v", itemID, userID, err)
			continue
		}
		totalSaved += saved
	}

	return totalSaved, nil
}

// upsertAccount insere ou atualiza uma conta conectada.
func (s *SyncService) upsertAccount(ctx context.Context, userID string, pa pluggy.Account, logo, color string) (string, error) {
	var id string
	
	err := s.db.QueryRow(ctx, `
		INSERT INTO connected_accounts (id, user_id, pluggy_item_id, pluggy_account_id, institution_name, institution_logo, institution_color, account_type, subtype, balance, currency)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (pluggy_account_id) DO NOTHING
		RETURNING id`, 
		userID, pa.ItemID, pa.ID, pa.MarketingName, logo, color, pa.Type, pa.Subtype, pa.Balance, pa.CurrencyCode).Scan(&id)

	if err != nil {
		// Se já existe ou conflito, buscamos o ID
		err = s.db.QueryRow(ctx, "SELECT id FROM connected_accounts WHERE pluggy_account_id = $1", pa.ID).Scan(&id)
		if err != nil {
			return "", err
		}
		// Atualiza o saldo e metadados (importante atualizar user_id caso a conta tenha mudado de dono no nosso sistema)
		s.db.Exec(ctx, `
			UPDATE connected_accounts 
			SET user_id = $1, balance = $2, account_type = $3, subtype = $4, institution_logo = $5, institution_color = $6, last_synced_at = NOW() 
			WHERE id = $7`, 
			userID, pa.Balance, pa.Type, pa.Subtype, logo, color, id)
	}

	return id, nil
}

// saveTransactionsAndReturn salva as transações e retorna as que foram inseridas/identificadas.
func (s *SyncService) saveTransactionsAndReturn(ctx context.Context, userID, accountID string, txs []pluggy.Transaction) ([]map[string]interface{}, error) {
	var saved []map[string]interface{}
	
	for _, tx := range txs {
		direction := "debit"
		if tx.Type == "CREDIT" {
			direction = "credit"
		}
		
		amount := tx.Amount
		if amount < 0 {
			amount = -amount
		}

		// Tenta classificar a transação
		categoryID := ""
		if s.classifierService != nil {
			catID, err := s.classifierService.Classify(ctx, userID, tx.Description, tx.Description, amount, direction)
			if err == nil {
				categoryID = catID
			}
		}

		var id string
		var err error
		if categoryID != "" {
			err = s.db.QueryRow(ctx, `
				INSERT INTO transactions (account_id, pluggy_transaction_id, amount, direction, description, merchant_name, date, category_id, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
				ON CONFLICT (pluggy_transaction_id) DO NOTHING
				RETURNING id`,
				accountID, tx.ID, amount, direction, tx.Description, tx.Description, tx.Date, categoryID).Scan(&id)
		} else {
			err = s.db.QueryRow(ctx, `
				INSERT INTO transactions (account_id, pluggy_transaction_id, amount, direction, description, merchant_name, date, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
				ON CONFLICT (pluggy_transaction_id) DO NOTHING
				RETURNING id`,
				accountID, tx.ID, amount, direction, tx.Description, tx.Description, tx.Date).Scan(&id)
		}
		
		if err == nil {
			txData := map[string]interface{}{
				"id":          id,
				"description": tx.Description,
				"amount":      amount,
				"direction":   direction,
				"date":        tx.Date,
			}
			if tx.CreditCardMetadata != nil {
				txData["installment_number"] = tx.CreditCardMetadata.InstallmentNumber
				txData["installments_count"] = tx.CreditCardMetadata.InstallmentsCount
			}
			saved = append(saved, txData)
		}
	}

	return saved, nil
}

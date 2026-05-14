package service

import (
	"context"
	"fmt"
	"log"
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
func (s *SyncService) SyncItem(ctx context.Context, userID string, itemID string, pluggyClient *pluggy.Client) error {
	// 1. Buscar detalhes do Item para pegar o ConnectorID
	item, err := pluggyClient.GetItem(itemID)
	var logo, color string
	if err == nil {
		// 2. Buscar detalhes do Conector (Logo e Cor)
		connector, err := pluggyClient.GetConnector(item.ConnectorID)
		if err == nil {
			logo = connector.ImageUrl
			color = connector.PrimaryColor
		}
	}

	// 3. Buscar contas do Item na Pluggy
	pluggyAccounts, err := pluggyClient.GetAccounts(itemID)
	if err != nil {
		return fmt.Errorf("erro ao buscar contas na pluggy: %w", err)
	}

	for _, pa := range pluggyAccounts {
		// 4. Upsert da conta no nosso banco com logo e cor
		accountID, err := s.upsertAccount(ctx, userID, pa, logo, color)
		if err != nil {
			log.Printf("erro ao sincronizar conta %s: %v", pa.ID, err)
			continue
		}

		// 3. Buscar transações dos últimos 90 dias
		to := time.Now().Format("2006-01-02")
		from := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
		
		transactions, err := pluggyClient.GetTransactions(pa.ID, from, to)
		if err != nil {
			log.Printf("erro ao buscar transações da conta %s: %v", pa.ID, err)
			continue
		}

		// 4. Salvar transações (passando userID para classificação)
		savedTxs, err := s.saveTransactionsAndReturn(ctx, userID, accountID, transactions)
		if err != nil {
			log.Printf("erro ao salvar transações da conta %s: %v", pa.ID, err)
		}

		// 5. Detectar parcelamentos
		if s.installmentService != nil && len(savedTxs) > 0 {
			if err := s.installmentService.ProcessTransactions(ctx, accountID, savedTxs); err != nil {
				log.Printf("erro ao processar parcelamentos: %v", err)
			}
		}

		// 6. Gerar eventos no feed
		if s.feedService != nil && len(savedTxs) > 0 {
			if err := s.feedService.GenerateEvents(ctx, userID, savedTxs); err != nil {
				log.Printf("erro ao gerar eventos no feed: %v", err)
			}
		}
		
		// 7. Atualizar last_synced_at
		s.db.Exec(ctx, "UPDATE connected_accounts SET last_synced_at = NOW() WHERE id = $1", accountID)
	}

	return nil
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

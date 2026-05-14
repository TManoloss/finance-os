package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedEventType string

const (
	EventDuplicateCharge    FeedEventType = "duplicate_charge"
	EventUnusualSpending    FeedEventType = "unusual_spending"
	EventSubscriptionChange FeedEventType = "subscription_change"
	EventNewMerchant        FeedEventType = "new_merchant"
	EventMilestone          FeedEventType = "milestone"
	EventInstallmentAlert   FeedEventType = "installment_alert"
	EventSalaryDetected     FeedEventType = "salary_detected"
	EventLowBalance         FeedEventType = "low_balance"
	EventMonthlyClose       FeedEventType = "monthly_close"
	EventAgentInsight       FeedEventType = "agent_insight"
)

type FeedEvent struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Type        FeedEventType  `json:"type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Amount      *float64       `json:"amount"`
	Severity    string         `json:"severity"` // info, warning, alert
	RelatedTx   []string       `json:"related_tx_ids"`
	ReadAt      *time.Time     `json:"read_at"`
	CreatedAt   time.Time      `json:"created_at"`
}

type FeedService struct {
	db *pgxpool.Pool
}

func NewFeedService(db *pgxpool.Pool) *FeedService {
	return &FeedService{db: db}
}

// GetFeed retorna os eventos do feed de um usuário.
func (s *FeedService) GetFeed(ctx context.Context, userID string, page, pageSize int) ([]FeedEvent, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, type, title, description, amount, severity, related_tx_ids, read_at, created_at
		FROM feed_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []FeedEvent
	for rows.Next() {
		var e FeedEvent
		err := rows.Scan(
			&e.ID, &e.UserID, &e.Type, &e.Title, &e.Description, &e.Amount,
			&e.Severity, &e.RelatedTx, &e.ReadAt, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// MarkAsRead marca um evento como lido.
func (s *FeedService) MarkAsRead(ctx context.Context, userID, eventID string) error {
	_, err := s.db.Exec(ctx, "UPDATE feed_events SET read_at = NOW() WHERE id = $1 AND user_id = $2", eventID, userID)
	return err
}

// MarkAllAsRead marca todos os eventos como lidos.
func (s *FeedService) MarkAllAsRead(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, "UPDATE feed_events SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL", userID)
	return err
}

// GetUnreadCount retorna o número de eventos não lidos.
func (s *FeedService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM feed_events WHERE user_id = $1 AND read_at IS NULL", userID).Scan(&count)
	return count, err
}

// GenerateEvents analisa transações e gera eventos no feed.
// Esta função deve ser chamada após uma sincronização.
func (s *FeedService) GenerateEvents(ctx context.Context, userID string, txs []map[string]interface{}) error {
	for _, tx := range txs {
		amount := tx["amount"].(float64)
		direction := tx["direction"].(string)
		description := tx["description"].(string)
		txID := tx["id"].(string)

		// 1. Detect Salary
		if direction == "credit" && amount >= 2000 {
			title := "Salário detectado! 💰"
			desc := "Recebemos um crédito de R$ " + formatAmount(amount) + " que parece ser sua renda principal."
			
			var exists bool
			s.db.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM feed_events 
					WHERE user_id = $1 AND type = $2 AND created_at > NOW() - INTERVAL '15 days'
				)
			`, userID, EventSalaryDetected).Scan(&exists)

			if !exists {
				s.CreateEvent(ctx, userID, EventSalaryDetected, title, desc, &amount, "info", []string{txID})
			}
		}

		// 2. Detect Large Unusual Spending
		if direction == "debit" && amount > 1000 {
			title := "Gasto elevado detectado ⚠️"
			desc := "Você teve um gasto de R$ " + formatAmount(amount) + " em " + description + ". Isso está acima do seu padrão habitual."
			s.CreateEvent(ctx, userID, EventUnusualSpending, title, desc, &amount, "warning", []string{txID})
		}

		// 3. Detect New Merchant
		var prevCount int
		s.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			WHERE a.user_id = $1 AND (t.merchant_name = $2 OR t.description = $2) AND t.id != $3
		`, userID, description, txID).Scan(&prevCount)

		if prevCount == 0 {
			title := "Novo estabelecimento 🛍️"
			desc := "Vimos que você comprou pela primeira vez em " + description + "."
			s.CreateEvent(ctx, userID, EventNewMerchant, title, desc, &amount, "info", []string{txID})
		}

		// 4. Detect Last Installment (Milestone)
		if instCount, ok := tx["installments_count"]; ok {
			instNum := tx["installment_number"].(int)
			count := instCount.(int)
			if instNum == count && count > 1 {
				title := "Parcelamento quitado! 🎉"
				desc := "Você pagou a última parcela (" + fmt.Sprintf("%d/%d", instNum, count) + ") de " + description + ". Menos uma conta!"
				s.CreateEvent(ctx, userID, EventMilestone, title, desc, &amount, "info", []string{txID})
			}
		}

		// 5. Detect Duplicate Charge
		var dupID string
		s.db.QueryRow(ctx, `
			SELECT t.id FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			WHERE a.user_id = $1 
			  AND (t.merchant_name = $2 OR t.description = $2) 
			  AND t.amount = $3 
			  AND t.id != $4
			  AND t.date >= (CAST($5 AS DATE) - INTERVAL '2 days')
			  AND t.date <= (CAST($5 AS DATE) + INTERVAL '2 days')
			LIMIT 1
		`, userID, description, amount, txID, tx["date"]).Scan(&dupID)

		if dupID != "" {
			title := "Possível cobrança duplicada 🔍"
			desc := "Detectamos dois gastos idênticos em " + description + " com valores de R$ " + formatAmount(amount) + " em datas próximas."
			
			var exists bool
			s.db.QueryRow(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM feed_events 
					WHERE user_id = $1 AND type = $2 AND (related_tx_ids @> ARRAY[$3]::uuid[] OR related_tx_ids @> ARRAY[$4]::uuid[])
				)
			`, userID, EventDuplicateCharge, txID, dupID).Scan(&exists)

			if !exists {
				s.CreateEvent(ctx, userID, EventDuplicateCharge, title, desc, &amount, "alert", []string{txID, dupID})
			}
		}
	}

	// 6. Detect Low Balance (Global check after all txs processed)
	var totalBalance float64
	s.db.QueryRow(ctx, "SELECT SUM(balance) FROM connected_accounts WHERE user_id = $1", userID).Scan(&totalBalance)

	if totalBalance < 500 {
		title := "Saldo em nível crítico 🚨"
		desc := "Seu saldo total consolidado está abaixo de R$ 500,00. Fique atento aos próximos compromissos."
		
		var exists bool
		s.db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM feed_events 
				WHERE user_id = $1 AND type = $2 AND created_at > NOW() - INTERVAL '3 days'
			)
		`, userID, EventLowBalance).Scan(&exists)

		if !exists {
			s.CreateEvent(ctx, userID, EventLowBalance, title, desc, &totalBalance, "alert", nil)
		}
	}

	return nil
}

func (s *FeedService) CreateEvent(ctx context.Context, userID string, eventType FeedEventType, title, description string, amount *float64, severity string, relatedTx []string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO feed_events (user_id, type, title, description, amount, severity, related_tx_ids)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, eventType, title, description, amount, severity, relatedTx)
	return err
}

func formatAmount(a float64) string {
	return string(fmt.Sprintf("%.2f", a))
}

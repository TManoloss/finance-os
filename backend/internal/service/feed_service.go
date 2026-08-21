package service

import (
	"context"
	"fmt"
	"math"
	"strings"
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
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	Type        FeedEventType `json:"type"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Amount      *float64      `json:"amount"`
	Severity    string        `json:"severity"` // info, warning, alert
	RelatedTx   []string      `json:"related_tx_ids"`
	ReadAt      *time.Time    `json:"read_at"`
	CreatedAt   time.Time     `json:"created_at"`
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
	var medianDebit float64
	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY t.amount), 0)
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= CURRENT_DATE - INTERVAL '90 days'
	`, userID).Scan(&medianDebit)
	highSpendingThreshold := unusualSpendingThreshold(medianDebit)

	for _, tx := range txs {
		amount := tx["amount"].(float64)
		direction := tx["direction"].(string)
		description := tx["description"].(string)
		txID := tx["id"].(string)

		// 1. Detect Salary
		isSalary := false
		if direction == "credit" {
			isSalary = isSalaryDescription(description)
			if !isSalary {
				_ = s.db.QueryRow(ctx, `
					SELECT COUNT(DISTINCT DATE_TRUNC('month', t.date)) >= 2
						AND COALESCE((MAX(t.amount) - MIN(t.amount)) / NULLIF(AVG(t.amount), 0), 1) <= 0.20
					FROM transactions t
					JOIN connected_accounts a ON t.account_id = a.id
					WHERE a.user_id = $1 AND t.direction = 'credit'
					  AND LOWER(COALESCE(NULLIF(t.merchant_name, ''), t.description)) = LOWER($2)
					  AND t.date >= DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '2 months'
				`, userID, description).Scan(&isSalary)
			}
		}
		if isSalary {
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
		if direction == "debit" && amount > highSpendingThreshold {
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
			  AND t.direction = 'debit'
			  AND LOWER(COALESCE(NULLIF(t.merchant_name, ''), t.description)) = LOWER($2)
			  AND t.amount = $3 
			  AND t.id != $4
			  AND ABS(t.date - CAST($5 AS DATE)) <= 2
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

	var nextIncome *time.Time
	var commitments float64
	_ = s.db.QueryRow(ctx, `
		WITH recurring_income AS (
			SELECT MAX(t.date) + INTERVAL '1 month' AS next_date
			FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			WHERE a.user_id = $1 AND t.direction = 'credit'
			  AND t.date >= DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '2 months'
			GROUP BY LOWER(COALESCE(NULLIF(t.merchant_name, ''), t.description))
			HAVING BOOL_OR(LOWER(t.description) ~ '(sal[aá]rio|folha de pagamento|pr[oó][ -]?labore)')
			   OR (COUNT(DISTINCT DATE_TRUNC('month', t.date)) >= 2
			       AND COALESCE((MAX(t.amount) - MIN(t.amount)) / NULLIF(AVG(t.amount), 0), 1) <= 0.20)
			ORDER BY MAX(t.date) DESC
			LIMIT 1
		), recurring_debits AS (
			SELECT AVG(t.amount) AS amount
			FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			CROSS JOIN recurring_income r
			WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= CURRENT_DATE - INTERVAL '90 days'
			GROUP BY LOWER(COALESCE(NULLIF(t.merchant_name, ''), t.description)), r.next_date
			HAVING COUNT(*) >= 2
			   AND (MAX(t.date) - MIN(t.date)) / NULLIF(COUNT(*) - 1, 0) BETWEEN 25 AND 35
			   AND COALESCE((MAX(t.amount) - MIN(t.amount)) / NULLIF(AVG(t.amount), 0), 1) <= 0.20
			   AND MAX(t.date) + ROUND((MAX(t.date) - MIN(t.date))::numeric / NULLIF(COUNT(*) - 1, 0))::int
			       BETWEEN CURRENT_DATE AND r.next_date
		), installment_commitments AS (
			SELECT COALESCE(SUM(i.total_amount / NULLIF(i.installments_total, 0)), 0) AS amount
			FROM installments i
			JOIN connected_accounts a ON i.account_id = a.id
			CROSS JOIN recurring_income r
			WHERE a.user_id = $1 AND i.next_due_date BETWEEN CURRENT_DATE AND r.next_date
		)
		SELECT r.next_date, COALESCE((SELECT SUM(amount) FROM recurring_debits), 0) + i.amount
		FROM recurring_income r CROSS JOIN installment_commitments i
	`, userID).Scan(&nextIncome, &commitments)

	criticalBalance := nextIncome == nil && totalBalance < 500
	if nextIncome != nil {
		criticalBalance = totalBalance < commitments
	}
	if criticalBalance {
		title := "Saldo em nível crítico 🚨"
		desc := "Seu saldo disponível está abaixo dos compromissos previstos até a próxima renda."
		if nextIncome == nil {
			desc = "Seu saldo total consolidado está abaixo de R$ 500,00; ainda não há histórico suficiente para projetar a próxima renda."
		}

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

func unusualSpendingThreshold(medianDebit float64) float64 {
	return math.Max(1000, medianDebit*3)
}

func isSalaryDescription(description string) bool {
	description = strings.ToLower(description)
	for _, term := range []string{"salário", "salario", "folha de pagamento", "pró-labore", "pro-labore", "pro labore"} {
		if strings.Contains(description, term) {
			return true
		}
	}
	return false
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

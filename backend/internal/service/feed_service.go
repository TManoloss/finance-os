package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var brlNumberPrinter = message.NewPrinter(language.BrazilianPortuguese)

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

		// 6. Detect Subscription Change (Mudança de valor em assinatura/recorrente)
		if direction == "debit" {
			var prevID string
			var prevAmount float64
			var prevDate time.Time
			err := s.db.QueryRow(ctx, `
				SELECT t.id, t.amount, t.date
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				WHERE a.user_id = $1
				  AND t.direction = 'debit'
				  AND LOWER(COALESCE(NULLIF(t.merchant_name, ''), t.description)) = LOWER($2)
				  AND t.id != $3
				  AND t.date BETWEEN CAST($4 AS DATE) - INTERVAL '45 days' AND CAST($4 AS DATE) - INTERVAL '20 days'
				ORDER BY t.date DESC
				LIMIT 1
			`, userID, description, txID, tx["date"]).Scan(&prevID, &prevAmount, &prevDate)

			if err == nil && prevID != "" && math.Abs(amount-prevAmount) >= 1.0 {
				var alreadyEmitted bool
				_ = s.db.QueryRow(ctx, `
					SELECT EXISTS(
						SELECT 1 FROM feed_events
						WHERE user_id = $1 AND type = $2
						  AND created_at > NOW() - INTERVAL '30 days'
						  AND (related_tx_ids @> ARRAY[$3]::uuid[] OR related_tx_ids @> ARRAY[$4]::uuid[])
					)
				`, userID, EventSubscriptionChange, txID, prevID).Scan(&alreadyEmitted)

				if !alreadyEmitted {
					title, desc, severity := formatSubscriptionChange(description, prevAmount, amount)
					s.CreateEvent(ctx, userID, EventSubscriptionChange, title, desc, &amount, severity, []string{txID, prevID})
				}
			}
		}
	}

	// 7. Detect Low Balance (Global check after all txs processed)
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

	// 8. Detect Monthly Close (Fechamento do mês anterior)
	var hasMonthlyClose bool
	_ = s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM feed_events
			WHERE user_id = $1 AND type = $2
			  AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
		)
	`, userID, EventMonthlyClose).Scan(&hasMonthlyClose)

	if !hasMonthlyClose {
		var totalSpent, totalIncome float64
		var txCount int
		err := s.db.QueryRow(ctx, `
			SELECT 
				COALESCE(SUM(CASE WHEN t.direction = 'debit' THEN t.amount ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN t.direction = 'credit' THEN t.amount ELSE 0 END), 0),
				COUNT(*)
			FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			WHERE a.user_id = $1
			  AND t.date >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
			  AND t.date < DATE_TRUNC('month', CURRENT_DATE)
		`, userID).Scan(&totalSpent, &totalIncome, &txCount)

		if err == nil && txCount > 0 {
			prevMonth := time.Now().AddDate(0, -1, 0)
			monthName := portugueseMonthName(prevMonth.Month())
			title, desc := formatMonthlyClose(monthName, totalSpent, totalIncome)
			s.CreateEvent(ctx, userID, EventMonthlyClose, title, desc, &totalSpent, "info", nil)
		}
	}

	// 9. Detect Category Spike Insight (Aumento expressivo por categoria vs mês anterior)
	var hasCategoryInsight bool
	_ = s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM feed_events
			WHERE user_id = $1 AND type = $2
			  AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
		)
	`, userID, EventAgentInsight).Scan(&hasCategoryInsight)

	if !hasCategoryInsight {
		var catName string
		var currTotal, prevTotal, growthPct float64
		var txIDs []string
		err := s.db.QueryRow(ctx, `
			WITH prev_month AS (
				SELECT t.category_id, c.name as category_name, SUM(t.amount) as prev_total
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				JOIN categories c ON t.category_id = c.id
				WHERE a.user_id = $1 AND t.direction = 'debit'
				  AND t.date >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
				  AND t.date < DATE_TRUNC('month', CURRENT_DATE)
				GROUP BY t.category_id, c.name
				HAVING SUM(t.amount) >= 100
			), curr_month AS (
				SELECT t.category_id, c.name as category_name, SUM(t.amount) as curr_total,
				       ARRAY_AGG(t.id::text) as tx_ids
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				JOIN categories c ON t.category_id = c.id
				WHERE a.user_id = $1 AND t.direction = 'debit'
				  AND t.date >= DATE_TRUNC('month', CURRENT_DATE)
				GROUP BY t.category_id, c.name
				HAVING SUM(t.amount) >= 150
			)
			SELECT curr.category_name, curr.curr_total, prev.prev_total,
			       ((curr.curr_total - prev.prev_total) / prev.prev_total) * 100 as growth_pct,
			       curr.tx_ids
			FROM curr_month curr
			JOIN prev_month prev ON curr.category_id = prev.category_id
			WHERE curr.curr_total >= prev.prev_total * 1.50
			ORDER BY (curr.curr_total - prev.prev_total) DESC
			LIMIT 1
		`, userID).Scan(&catName, &currTotal, &prevTotal, &growthPct, &txIDs)

		if err == nil && catName != "" {
			title, desc := formatCategorySpike(catName, currTotal, prevTotal, growthPct)
			limitIDs := txIDs
			if len(limitIDs) > 5 {
				limitIDs = limitIDs[:5]
			}
			s.CreateEvent(ctx, userID, EventAgentInsight, title, desc, &currTotal, "warning", limitIDs)
		}
	}

	return nil
}

func formatSubscriptionChange(merchant string, oldAmount, newAmount float64) (title, desc, severity string) {
	diff := newAmount - oldAmount
	if diff > 0 {
		return "Aumento na assinatura 💳",
			fmt.Sprintf("A cobrança de %s passou de R$ %s para R$ %s (+R$ %s).", merchant, formatAmount(oldAmount), formatAmount(newAmount), formatAmount(diff)),
			"warning"
	}
	return "Redução na assinatura 🎉",
		fmt.Sprintf("A cobrança de %s reduziu de R$ %s para R$ %s (-R$ %s).", merchant, formatAmount(oldAmount), formatAmount(newAmount), formatAmount(-diff)),
		"info"
}

func formatMonthlyClose(monthName string, totalSpent, totalIncome float64) (title, desc string) {
	net := totalIncome - totalSpent
	sign := ""
	if net > 0 {
		sign = "+"
	}
	return fmt.Sprintf("Fechamento de %s 📊", monthName),
		fmt.Sprintf("No mês anterior, você teve R$ %s em gastos e R$ %s em entradas (resultado líquido: %sR$ %s).",
			formatAmount(totalSpent), formatAmount(totalIncome), sign, formatAmount(net))
}

func formatCategorySpike(categoryName string, currTotal, prevTotal, growthPct float64) (title, desc string) {
	return fmt.Sprintf("Aumento em %s 📈", categoryName),
		fmt.Sprintf("Seus gastos com %s este mês (R$ %s) já superam em %.0f%% o total do mês anterior (R$ %s).",
			categoryName, formatAmount(currTotal), growthPct, formatAmount(prevTotal))
}

func portugueseMonthName(m time.Month) string {
	months := map[time.Month]string{
		time.January:   "Janeiro",
		time.February:  "Fevereiro",
		time.March:     "Março",
		time.April:     "Abril",
		time.May:       "Maio",
		time.June:      "Junho",
		time.July:      "Julho",
		time.August:    "Agosto",
		time.September: "Setembro",
		time.October:   "Outubro",
		time.November:  "Novembro",
		time.December:  "Dezembro",
	}
	return months[m]
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
	return brlNumberPrinter.Sprintf("%.2f", a)
}

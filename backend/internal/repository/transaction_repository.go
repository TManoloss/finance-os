package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionFilters define os filtros para listagem de transações.
type TransactionFilters struct {
	UserID     string
	AccountID  string
	CategoryID string
	FromDate   time.Time
	ToDate     time.Time
	Direction  string // debit, credit
	Page       int
	PageSize   int
}

// TransactionSummary representa o resumo financeiro de um período.
type TransactionSummary struct {
	TotalSpent      float64           `json:"total_spent"`
	TotalReceived   float64           `json:"total_received"`
	CheckingBalance float64           `json:"checking_balance"`
	CreditBalance   float64           `json:"credit_balance"`
	CurrentInvoice  float64           `json:"current_invoice"`  // Fatura acumulando (em aberto)
	ClosedInvoice   float64           `json:"closed_invoice"`   // Fatura já fechada (vencimento próximo)
	MonthInstallments float64         `json:"month_installments"` // Soma das parcelas que vencem este mês
	TodaySpent      float64           `json:"today_spent"`
	WeeklySpent     float64           `json:"weekly_spent"`
	ByCategory      []CategorySummary `json:"by_category"`
	ByDay           []DaySummary      `json:"by_day"`
	TopMerchants    []MerchantSummary `json:"top_merchants"`
}

type CategorySummary struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Color        string  `json:"color"`
	Total        float64 `json:"total"`
	Percentage   float64 `json:"percentage"`
	Count        int     `json:"transaction_count"`
}

type DaySummary struct {
	Date          time.Time `json:"date"`
	TotalSpent    float64   `json:"total_spent"`
	TotalReceived float64   `json:"total_received"`
}

type MerchantSummary struct {
	MerchantName string  `json:"merchant_name"`
	Total        float64 `json:"total"`
	Count        int     `json:"count"`
}

// TransactionRepository define a interface para persistência de transações.
type TransactionRepository interface {
	GetTransactions(ctx context.Context, filters TransactionFilters) ([]map[string]interface{}, int, error)
	GetSummary(ctx context.Context, userID string, from, to time.Time) (*TransactionSummary, error)
	UpdateCategory(ctx context.Context, txID, categoryID string) error
}

type pgTransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &pgTransactionRepository{db: db}
}

func (r *pgTransactionRepository) GetTransactions(ctx context.Context, f TransactionFilters) ([]map[string]interface{}, int, error) {
	offset := (f.Page - 1) * f.PageSize

	query := `
		SELECT 
			t.id, t.amount, t.direction, t.description, t.merchant_name, t.date, t.is_recurring,
			c.id as category_id, c.name as category_name, c.color as category_color,
			acc.institution_name as account_name
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE acc.user_id = $1
	`
	args := []interface{}{f.UserID}
	argCount := 2

	if f.AccountID != "" {
		query += fmt.Sprintf(" AND t.account_id = $%d", argCount)
		args = append(args, f.AccountID)
		argCount++
	}
	if f.CategoryID != "" {
		query += fmt.Sprintf(" AND t.category_id = $%d", argCount)
		args = append(args, f.CategoryID)
		argCount++
	}
	if !f.FromDate.IsZero() {
		query += fmt.Sprintf(" AND t.date >= $%d", argCount)
		args = append(args, f.FromDate)
		argCount++
	}
	if !f.ToDate.IsZero() {
		query += fmt.Sprintf(" AND t.date <= $%d", argCount)
		args = append(args, f.ToDate)
		argCount++
	}
	if f.Direction != "" {
		query += fmt.Sprintf(" AND t.direction = $%d", argCount)
		args = append(args, f.Direction)
		argCount++
	}

	// Total count para paginação
	countQuery := "SELECT COUNT(*) FROM (" + query + ") as total"
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Ordenação e paginação
	query += fmt.Sprintf(" ORDER BY t.date DESC, t.created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	transactions := []map[string]interface{}{}
	for rows.Next() {
		var tx struct {
			ID            string
			Amount        float64
			Direction     string
			Description   string
			MerchantName  *string
			Date          time.Time
			IsRecurring   bool
			CategoryID    *string
			CategoryName  *string
			CategoryColor *string
			AccountName   string
		}
		err := rows.Scan(
			&tx.ID, &tx.Amount, &tx.Direction, &tx.Description, &tx.MerchantName, &tx.Date, &tx.IsRecurring,
			&tx.CategoryID, &tx.CategoryName, &tx.CategoryColor, &tx.AccountName,
		)
		if err != nil {
			return nil, 0, err
		}

		item := map[string]interface{}{
			"id":            tx.ID,
			"amount":        tx.Amount,
			"direction":     tx.Direction,
			"description":   tx.Description,
			"merchant_name": tx.MerchantName,
			"date":          tx.Date,
			"is_recurring":  tx.IsRecurring,
			"account_name":  tx.AccountName,
			"category":      nil,
		}

		if tx.CategoryID != nil {
			item["category"] = map[string]interface{}{
				"id":    *tx.CategoryID,
				"name":  *tx.CategoryName,
				"color": *tx.CategoryColor,
			}
		}

		transactions = append(transactions, item)
	}

	return transactions, total, nil
}

func (r *pgTransactionRepository) UpdateCategory(ctx context.Context, txID, categoryID string) error {
	query := `UPDATE transactions SET category_id = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, categoryID, txID)
	return err
}

func (r *pgTransactionRepository) GetSummary(ctx context.Context, userID string, from, to time.Time) (*TransactionSummary, error) {
	summary := &TransactionSummary{
		ByCategory:   []CategorySummary{},
		ByDay:        []DaySummary{},
		TopMerchants: []MerchantSummary{},
	}

	// 1. Totais e Por Categoria
	categoryQuery := `
		WITH totals AS (
			SELECT SUM(amount) as total_spent
			FROM transactions t
			JOIN connected_accounts acc ON t.account_id = acc.id
			WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date BETWEEN $2 AND $3
		)
		SELECT 
			c.id, c.name, c.color, 
			SUM(t.amount) as total,
			COUNT(t.id) as count,
			(SUM(t.amount) / NULLIF((SELECT total_spent FROM totals), 0)) * 100 as percentage
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		JOIN categories c ON t.category_id = c.id
		WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date BETWEEN $2 AND $3
		GROUP BY c.id, c.name, c.color
		ORDER BY total DESC
	`
	rows, err := r.db.Query(ctx, categoryQuery, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cat CategorySummary
		if err := rows.Scan(&cat.CategoryID, &cat.CategoryName, &cat.Color, &cat.Total, &cat.Count, &cat.Percentage); err != nil {
			return nil, err
		}
		summary.ByCategory = append(summary.ByCategory, cat)
	}

	// 2. Totais Gerais (Excluindo transferências internas e ajustando por tipo de conta)
	// 'spent' = Débitos em qualquer conta (exceto pagamentos de fatura/transferências)
	// 'received' = Créditos APENAS em contas bancárias (ignora créditos em cartão como pagamento de fatura)
	totalsQuery := `
		SELECT 
			COALESCE(SUM(CASE 
				WHEN t.direction = 'debit' 
				AND t.description NOT ILIKE '%PAGAMENTO%FATURA%' 
				AND t.description NOT ILIKE '%TRANSFERENCIA%ENVIADA%' 
				AND t.description NOT ILIKE '%APLICAÇÃO%' 
				THEN t.amount ELSE 0 END), 0) as spent,
			COALESCE(SUM(CASE 
				WHEN t.direction = 'credit' 
				AND acc.account_type IN ('CHECKING', 'SAVINGS', 'BANK', 'checking', 'savings')
				AND t.description NOT ILIKE '%PAGAMENTO%FATURA%' 
				AND t.description NOT ILIKE '%TRANSFERENCIA%RECEBIDA%' 
				AND t.description NOT ILIKE '%RESGATE%' 
				THEN t.amount ELSE 0 END), 0) as received
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 AND t.date BETWEEN $2 AND $3
	`
	err = r.db.QueryRow(ctx, totalsQuery, userID, from, to).Scan(&summary.TotalSpent, &summary.TotalReceived)
	if err != nil {
		return nil, err
	}

	// 3. Por Dia (Seguindo a mesma lógica de filtro)
	dayQuery := `
		SELECT 
			t.date,
			SUM(CASE 
				WHEN t.direction = 'debit' 
				AND t.description NOT ILIKE '%PAGAMENTO%FATURA%' 
				AND t.description NOT ILIKE '%TRANSFERENCIA%ENVIADA%' 
				AND t.description NOT ILIKE '%APLICAÇÃO%' 
				THEN t.amount ELSE 0 END) as spent,
			SUM(CASE 
				WHEN t.direction = 'credit' 
				AND acc.account_type IN ('CHECKING', 'SAVINGS', 'BANK', 'checking', 'savings')
				AND t.description NOT ILIKE '%PAGAMENTO%FATURA%' 
				AND t.description NOT ILIKE '%TRANSFERENCIA%RECEBIDA%' 
				AND t.description NOT ILIKE '%RESGATE%' 
				THEN t.amount ELSE 0 END) as received
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 AND t.date BETWEEN $2 AND $3
		GROUP BY t.date
		ORDER BY t.date ASC
	`
	rows, err = r.db.Query(ctx, dayQuery, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day DaySummary
		if err := rows.Scan(&day.Date, &day.TotalSpent, &day.TotalReceived); err != nil {
			return nil, err
		}
		summary.ByDay = append(summary.ByDay, day)
	}

	// 4. Top Merchants
	merchantQuery := `
		SELECT 
			COALESCE(merchant_name, description) as merchant,
			SUM(amount) as total,
			COUNT(*) as count
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date BETWEEN $2 AND $3
		AND description NOT ILIKE '%PAGAMENTO%FATURA%' AND description NOT ILIKE '%TRANSFERENCIA%ENVIADA%' AND description NOT ILIKE '%APLICAÇÃO%'
		GROUP BY merchant
		ORDER BY total DESC
		LIMIT 5
	`
	rows, err = r.db.Query(ctx, merchantQuery, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m MerchantSummary
		if err := rows.Scan(&m.MerchantName, &m.Total, &m.Count); err != nil {
			return nil, err
		}
		summary.TopMerchants = append(summary.TopMerchants, m)
	}

	// 5. Saldos por Tipo de Conta
	// Importante: Crédito deve ser exibido como valor positivo na UI (dívida),
	// mas no saldo total do patrimônio líquido ele deve ser subtraído.
	balanceQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN account_type IN ('CHECKING', 'SAVINGS', 'BANK', 'checking', 'savings') THEN balance ELSE 0 END), 0) as checking,
			COALESCE(SUM(CASE WHEN account_type IN ('CREDIT', 'credit') THEN balance ELSE 0 END), 0) as credit
		FROM connected_accounts
		WHERE user_id = $1
	`
	err = r.db.QueryRow(ctx, balanceQuery, userID).Scan(&summary.CheckingBalance, &summary.CreditBalance)
	if err != nil {
		return nil, err
	}

	// 6. Cálculo de Fatura Atual (Em Aberto) e Fatura Fechada
	// Fatura em Aberto = Gastos após o último fechamento até hoje
	invoiceQuery := `
		SELECT 
			COALESCE(SUM(CASE 
				WHEN t.date > (
					CASE 
						WHEN EXTRACT(DAY FROM CURRENT_DATE) >= COALESCE(acc.close_day, 1) 
						THEN DATE_TRUNC('month', CURRENT_DATE) + (COALESCE(acc.close_day, 1) - 1) * INTERVAL '1 day'
						ELSE DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') + (COALESCE(acc.close_day, 1) - 1) * INTERVAL '1 day'
					END
				) THEN t.amount ELSE 0 END), 0) as open_invoice,
			COALESCE(SUM(CASE 
				WHEN t.date <= (
					CASE 
						WHEN EXTRACT(DAY FROM CURRENT_DATE) >= COALESCE(acc.close_day, 1) 
						THEN DATE_TRUNC('month', CURRENT_DATE) + (COALESCE(acc.close_day, 1) - 1) * INTERVAL '1 day'
						ELSE DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') + (COALESCE(acc.close_day, 1) - 1) * INTERVAL '1 day'
					END
				) 
				AND t.date > (
					CASE 
						WHEN EXTRACT(DAY FROM CURRENT_DATE) >= COALESCE(acc.close_day, 1) 
						THEN DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') + (COALESCE(acc.close_day, 1) - 1) * INTERVAL '1 day'
						ELSE DATE_TRUNC('month', CURRENT_DATE - INTERVAL '2 month') + (COALESCE(acc.close_day, 1) - 1) * INTERVAL '1 day'
					END
				) THEN t.amount ELSE 0 END), 0) as closed_invoice
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 
		AND acc.account_type IN ('CREDIT', 'credit')
		AND t.direction = 'debit'
		AND t.description NOT ILIKE '%PAGAMENTO%FATURA%'
	`
	err = r.db.QueryRow(ctx, invoiceQuery, userID).Scan(&summary.CurrentInvoice, &summary.ClosedInvoice)
	if err != nil {
		return nil, err
	}

	// 7. Soma das Parcelas do Mês Atual
	installmentsQuery := `
		SELECT COALESCE(SUM(total_amount / installments_total), 0)
		FROM installments
		WHERE account_id IN (SELECT id FROM connected_accounts WHERE user_id = $1)
		AND next_due_date BETWEEN DATE_TRUNC('month', CURRENT_DATE) AND (DATE_TRUNC('month', CURRENT_DATE) + INTERVAL '1 month' - INTERVAL '1 day')
	`
	err = r.db.QueryRow(ctx, installmentsQuery, userID).Scan(&summary.MonthInstallments)
	if err != nil {
		// Se a tabela não existir ou outro erro, apenas ignoramos para não quebrar o dashboard
		log.Printf("Aviso: erro ao buscar parcelas: %v", err)
	}

	// 8. Gastos de Hoje e da Semana
	dailyWeeklyQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN t.date = $2 AND t.direction = 'debit' AND t.description NOT ILIKE '%PAGAMENTO%FATURA%' AND t.description NOT ILIKE '%APLICAÇÃO%' THEN t.amount ELSE 0 END), 0) as today,
			COALESCE(SUM(CASE WHEN t.date > $2 - INTERVAL '7 days' AND t.direction = 'debit' AND t.description NOT ILIKE '%PAGAMENTO%FATURA%' AND t.description NOT ILIKE '%APLICAÇÃO%' THEN t.amount ELSE 0 END), 0) as weekly
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1
	`
	// Usamos o dia atual do servidor Go para garantir consistência de timezone
	today := time.Now().Format("2006-01-02")
	err = r.db.QueryRow(ctx, dailyWeeklyQuery, userID, today).Scan(&summary.TodaySpent, &summary.WeeklySpent)
	if err != nil {
		log.Printf("Aviso: erro ao buscar gastos diários/semanais: %v", err)
	}

	return summary, nil
}

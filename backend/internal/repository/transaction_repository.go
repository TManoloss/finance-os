package repository

import (
	"context"
	"fmt"
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

	var transactions []map[string]interface{}
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
			(SUM(t.amount) / (SELECT total_spent FROM totals)) * 100 as percentage
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

	// 2. Totais Gerais
	totalsQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN direction = 'debit' THEN amount ELSE 0 END), 0) as spent,
			COALESCE(SUM(CASE WHEN direction = 'credit' THEN amount ELSE 0 END), 0) as received
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 AND t.date BETWEEN $2 AND $3
	`
	err = r.db.QueryRow(ctx, totalsQuery, userID, from, to).Scan(&summary.TotalSpent, &summary.TotalReceived)
	if err != nil {
		return nil, err
	}

	// 3. Por Dia
	dayQuery := `
		SELECT 
			date,
			SUM(CASE WHEN direction = 'debit' THEN amount ELSE 0 END) as spent,
			SUM(CASE WHEN direction = 'credit' THEN amount ELSE 0 END) as received
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 AND t.date BETWEEN $2 AND $3
		GROUP BY date
		ORDER BY date ASC
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
	balanceQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN account_type IN ('checking', 'savings') THEN balance ELSE 0 END), 0) as checking,
			COALESCE(SUM(CASE WHEN account_type = 'credit' THEN balance ELSE 0 END), 0) as credit
		FROM connected_accounts
		WHERE user_id = $1
	`
	err = r.db.QueryRow(ctx, balanceQuery, userID).Scan(&summary.CheckingBalance, &summary.CreditBalance)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

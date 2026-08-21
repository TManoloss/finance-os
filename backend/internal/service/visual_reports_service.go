package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MonthlyReplayResult struct {
	Month                    string                 `json:"month"`
	MonthName                string                 `json:"month_name"`
	TotalSpent               float64                `json:"total_spent"`
	TotalReceived            float64                `json:"total_received"`
	NetSavings               float64                `json:"net_savings"`
	TopTransaction           map[string]interface{} `json:"top_transaction,omitempty"`
	TopMerchant              map[string]interface{} `json:"top_merchant,omitempty"`
	TopCategory              map[string]interface{} `json:"top_category,omitempty"`
	SpendingEvolutionPercent float64                `json:"spending_evolution_percent"`
	Insights                 []string               `json:"insights"`
}

type VisualReportsService struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewVisualReportsService(db *pgxpool.Pool, cfg *config.Config) *VisualReportsService {
	return &VisualReportsService{db: db, cfg: cfg}
}

// GetSpendingHeatmap returns daily spending totals for the last 365 days
func (s *VisualReportsService) GetSpendingHeatmap(userID string) (interface{}, error) {
	query := `
		SELECT t.date, SUM(t.amount) as total
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= CURRENT_DATE - INTERVAL '365 days'
		GROUP BY t.date
		ORDER BY t.date ASC
	`

	rows, err := s.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type HeatmapDay struct {
		Date  string  `json:"date"`
		Total float64 `json:"total"`
	}

	var heatmap []HeatmapDay
	for rows.Next() {
		var date time.Time
		var total float64
		if err := rows.Scan(&date, &total); err != nil {
			return nil, err
		}
		heatmap = append(heatmap, HeatmapDay{
			Date:  date.Format("2006-01-02"),
			Total: total,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return heatmap, nil
}

func (s *VisualReportsService) GetDependencyMap(userID string) (interface{}, error) {
	// Mapa de dependência real derivado de merchants dos últimos 90 dias
	rows, err := s.db.Query(context.Background(), `
		SELECT COALESCE(t.merchant_name, 'Outros') as merchant, SUM(t.amount) as total, COUNT(*) as count
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= CURRENT_DATE - INTERVAL '90 days'
		GROUP BY merchant
		ORDER BY total DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type MerchantDep struct {
		Name  string  `json:"name"`
		Total float64 `json:"total"`
		Count int     `json:"count"`
	}
	var deps []MerchantDep
	for rows.Next() {
		var d MerchantDep
		if err := rows.Scan(&d.Name, &d.Total, &d.Count); err == nil {
			deps = append(deps, d)
		}
	}
	return deps, nil
}

// GetMonthlyReplay gera o replay financeiro do mês com dados reais (FOS-706).
func (s *VisualReportsService) GetMonthlyReplay(userID, month string) (*MonthlyReplayResult, error) {
	ctx := context.Background()
	parsedDate, err := time.Parse("2006-01", month)
	if err != nil {
		parsedDate = time.Now()
		month = parsedDate.Format("2006-01")
	}

	ptMonths := []string{"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}
	monthName := fmt.Sprintf("%s de %d", ptMonths[parsedDate.Month()], parsedDate.Year())

	startOfMonth := time.Date(parsedDate.Year(), parsedDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	// 1. Totais do mês
	var totalSpent, totalReceived float64
	_ = s.db.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN t.direction = 'debit' THEN t.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN t.direction = 'credit' THEN t.amount ELSE 0 END), 0)
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.date >= $2 AND t.date <= $3
	`, userID, startOfMonth, endOfMonth).Scan(&totalSpent, &totalReceived)

	netSavings := totalReceived - totalSpent

	// 2. Maior compra do mês
	var topTxDesc, topTxMerchant, topTxCategory string
	var topTxAmount float64
	var topTxDate time.Time
	topTxMap := make(map[string]interface{})

	err = s.db.QueryRow(ctx, `
		SELECT t.description, COALESCE(t.merchant_name, ''), COALESCE(c.name, 'Geral'), t.amount, t.date
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date <= $3
		ORDER BY t.amount DESC
		LIMIT 1
	`, userID, startOfMonth, endOfMonth).Scan(&topTxDesc, &topTxMerchant, &topTxCategory, &topTxAmount, &topTxDate)

	if err == nil && topTxAmount > 0 {
		topTxMap["description"] = topTxDesc
		topTxMap["merchant_name"] = topTxMerchant
		topTxMap["category_name"] = topTxCategory
		topTxMap["amount"] = topTxAmount
		topTxMap["date"] = topTxDate.Format("02/01/2006")
	}

	// 3. Principal estabelecimento do mês
	var topMName string
	var topMTotal float64
	var topMCount int
	topMMap := make(map[string]interface{})

	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(t.merchant_name, 'Outros'), SUM(t.amount) as total, COUNT(*) as count
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date <= $3
		GROUP BY t.merchant_name
		ORDER BY total DESC
		LIMIT 1
	`, userID, startOfMonth, endOfMonth).Scan(&topMName, &topMTotal, &topMCount)

	if err == nil && topMTotal > 0 {
		topMMap["merchant_name"] = topMName
		topMMap["total_spent"] = topMTotal
		topMMap["count"] = topMCount
	}

	// 4. Principal categoria do mês
	var topCatName string
	var topCatTotal float64
	topCatMap := make(map[string]interface{})

	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(c.name, 'Outros'), SUM(t.amount) as total
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date <= $3
		GROUP BY c.name
		ORDER BY total DESC
		LIMIT 1
	`, userID, startOfMonth, endOfMonth).Scan(&topCatName, &topCatTotal)

	if err == nil && topCatTotal > 0 {
		topCatMap["category_name"] = topCatName
		topCatMap["total_spent"] = topCatTotal
	}

	// 5. Comparativo com o mês anterior
	startOfPrevMonth := startOfMonth.AddDate(0, -1, 0)
	endOfPrevMonth := startOfMonth.AddDate(0, 0, -1)

	var prevMonthSpent float64
	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(t.amount), 0)
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date <= $3
	`, userID, startOfPrevMonth, endOfPrevMonth).Scan(&prevMonthSpent)

	var spendingEvolutionPercent float64
	if prevMonthSpent > 0 {
		spendingEvolutionPercent = ((totalSpent - prevMonthSpent) / prevMonthSpent) * 100
	}

	// 6. Insights determinísticos
	var insights []string
	if netSavings >= 0 {
		insights = append(insights, fmt.Sprintf("Mês com saldo positivo: você fechou com sobra de R$ %.2f.", netSavings))
	} else {
		insights = append(insights, fmt.Sprintf("Mês com déficit de R$ %.2f em relação às entradas.", math.Abs(netSavings)))
	}

	if prevMonthSpent > 0 {
		if spendingEvolutionPercent > 0 {
			insights = append(insights, fmt.Sprintf("Gastos aumentaram %.1f%% em relação ao mês anterior.", spendingEvolutionPercent))
		} else {
			insights = append(insights, fmt.Sprintf("Gastos reduziram %.1f%% em relação ao mês anterior.", math.Abs(spendingEvolutionPercent)))
		}
	}

	if topCatName != "" {
		insights = append(insights, fmt.Sprintf("Sua principal área de despesas foi %s (R$ %.2f).", topCatName, topCatTotal))
	}

	return &MonthlyReplayResult{
		Month:                    month,
		MonthName:                monthName,
		TotalSpent:               totalSpent,
		TotalReceived:            totalReceived,
		NetSavings:               netSavings,
		TopTransaction:           topTxMap,
		TopMerchant:              topMMap,
		TopCategory:              topCatMap,
		SpendingEvolutionPercent: spendingEvolutionPercent,
		Insights:                 insights,
	}, nil
}

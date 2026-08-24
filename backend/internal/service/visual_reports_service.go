package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MonthlyReplayResult struct {
	Month                    string                 `json:"month"`
	MonthName                string                 `json:"month_name"`
	HasData                  bool                   `json:"has_data"`
	HasPreviousSpending      bool                   `json:"has_previous_spending"`
	TotalSpent               float64                `json:"total_spent"`
	TotalReceived            float64                `json:"total_received"`
	NetSavings               float64                `json:"net_savings"`
	TopTransaction           map[string]interface{} `json:"top_transaction,omitempty"`
	TopMerchant              map[string]interface{} `json:"top_merchant,omitempty"`
	TopCategory              map[string]interface{} `json:"top_category,omitempty"`
	GrowingCategory          map[string]interface{} `json:"growing_category,omitempty"`
	SpendingEvolutionPercent float64                `json:"spending_evolution_percent"`
	MonthlyOutcome           string                 `json:"monthly_outcome"`
	NextMonthGuidance        string                 `json:"next_month_guidance"`
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
func (s *VisualReportsService) GetMonthlyReplay(ctx context.Context, userID, month string) (*MonthlyReplayResult, error) {
	parsedDate, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, fmt.Errorf("invalid replay month: %w", err)
	}

	ptMonths := []string{"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}
	monthName := fmt.Sprintf("%s de %d", ptMonths[parsedDate.Month()], parsedDate.Year())

	startOfMonth := time.Date(parsedDate.Year(), parsedDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)

	// 1. Totais do mês
	var totalSpent, totalReceived float64
	if err := s.db.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN t.direction = 'debit' THEN t.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN t.direction = 'credit' THEN t.amount ELSE 0 END), 0)
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.date >= $2 AND t.date < $3
	`, userID, startOfMonth, startOfNextMonth).Scan(&totalSpent, &totalReceived); err != nil {
		return nil, err
	}

	netSavings := totalReceived - totalSpent
	hasData := totalSpent > 0 || totalReceived > 0

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
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date < $3
		ORDER BY t.amount DESC
		LIMIT 1
	`, userID, startOfMonth, startOfNextMonth).Scan(&topTxDesc, &topTxMerchant, &topTxCategory, &topTxAmount, &topTxDate)

	if err == nil && topTxAmount > 0 {
		topTxMap["description"] = topTxDesc
		topTxMap["merchant_name"] = topTxMerchant
		topTxMap["category_name"] = topTxCategory
		topTxMap["amount"] = topTxAmount
		topTxMap["date"] = topTxDate.Format("02/01/2006")
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
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
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date < $3
		GROUP BY t.merchant_name
		ORDER BY total DESC
		LIMIT 1
	`, userID, startOfMonth, startOfNextMonth).Scan(&topMName, &topMTotal, &topMCount)

	if err == nil && topMTotal > 0 {
		topMMap["merchant_name"] = topMName
		topMMap["total_spent"] = topMTotal
		topMMap["count"] = topMCount
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 4. Principal categoria do mês
	var topCatName string
	var topCatTotal float64
	topCatMap := make(map[string]interface{})

	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(parent.name, c.name, 'Outros'), SUM(t.amount) as total
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		LEFT JOIN categories c ON t.category_id = c.id
		LEFT JOIN categories parent ON c.parent_id = parent.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date < $3
		GROUP BY COALESCE(parent.name, c.name, 'Outros')
		ORDER BY total DESC
		LIMIT 1
	`, userID, startOfMonth, startOfNextMonth).Scan(&topCatName, &topCatTotal)

	if err == nil && topCatTotal > 0 {
		topCatMap["category_name"] = topCatName
		topCatMap["total_spent"] = topCatTotal
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// 5. Comparativo com o mês anterior
	startOfPrevMonth := startOfMonth.AddDate(0, -1, 0)

	var prevMonthSpent float64
	if err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(t.amount), 0)
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date < $3
	`, userID, startOfPrevMonth, startOfMonth).Scan(&prevMonthSpent); err != nil {
		return nil, err
	}

	var spendingEvolutionPercent float64
	if prevMonthSpent > 0 {
		spendingEvolutionPercent = ((totalSpent - prevMonthSpent) / prevMonthSpent) * 100
	}

	// 6. Categoria com maior crescimento absoluto contra o mês anterior.
	var growingName string
	var growingCurrent, growingPrevious, growingAmount float64
	growingCategory := make(map[string]interface{})
	if prevMonthSpent > 0 {
		err = s.db.QueryRow(ctx, `
		WITH current_period AS (
			SELECT COALESCE(parent.name, c.name, 'Outros') AS category, SUM(t.amount) AS total
			FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			LEFT JOIN categories c ON t.category_id = c.id
			LEFT JOIN categories parent ON c.parent_id = parent.id
			WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2 AND t.date < $3
			GROUP BY COALESCE(parent.name, c.name, 'Outros')
		), previous_period AS (
			SELECT COALESCE(parent.name, c.name, 'Outros') AS category, SUM(t.amount) AS total
			FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			LEFT JOIN categories c ON t.category_id = c.id
			LEFT JOIN categories parent ON c.parent_id = parent.id
			WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $4 AND t.date < $2
			GROUP BY COALESCE(parent.name, c.name, 'Outros')
		)
		SELECT current_period.category, current_period.total,
			COALESCE(previous_period.total, 0),
			current_period.total - COALESCE(previous_period.total, 0) AS growth
		FROM current_period
		LEFT JOIN previous_period USING (category)
		WHERE current_period.total > COALESCE(previous_period.total, 0)
		ORDER BY growth DESC
		LIMIT 1
		`, userID, startOfMonth, startOfNextMonth, startOfPrevMonth).Scan(
			&growingName, &growingCurrent, &growingPrevious, &growingAmount,
		)
		if err == nil {
			growingCategory["category_name"] = growingName
			growingCategory["current_spent"] = growingCurrent
			growingCategory["previous_spent"] = growingPrevious
			growingCategory["growth_amount"] = growingAmount
			if growingPrevious > 0 {
				growingCategory["growth_percent"] = growingAmount / growingPrevious * 100
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	// 7. Resultado e orientação determinísticos.
	monthlyOutcome, nextMonthGuidance := replayOutcome(netSavings, totalSpent, prevMonthSpent, growingName)
	var insights []string
	if !hasData {
		insights = append(insights, "Nenhuma movimentação encontrada para este mês.")
	} else if netSavings >= 0 {
		insights = append(insights, fmt.Sprintf("Mês com saldo positivo: você fechou com sobra de R$ %s.", formatAmount(netSavings)))
	} else {
		insights = append(insights, fmt.Sprintf("Mês com déficit de R$ %s em relação às entradas.", formatAmount(math.Abs(netSavings))))
	}

	if prevMonthSpent > 0 {
		if spendingEvolutionPercent > 0 {
			insights = append(insights, fmt.Sprintf("Gastos aumentaram %.1f%% em relação ao mês anterior.", spendingEvolutionPercent))
		} else {
			insights = append(insights, fmt.Sprintf("Gastos reduziram %.1f%% em relação ao mês anterior.", math.Abs(spendingEvolutionPercent)))
		}
	}

	if topCatName != "" {
		insights = append(insights, fmt.Sprintf("Sua principal área de despesas foi %s (R$ %s).", topCatName, formatAmount(topCatTotal)))
	}

	return &MonthlyReplayResult{
		Month:                    month,
		MonthName:                monthName,
		HasData:                  hasData,
		HasPreviousSpending:      prevMonthSpent > 0,
		TotalSpent:               totalSpent,
		TotalReceived:            totalReceived,
		NetSavings:               netSavings,
		TopTransaction:           topTxMap,
		TopMerchant:              topMMap,
		TopCategory:              topCatMap,
		GrowingCategory:          growingCategory,
		SpendingEvolutionPercent: spendingEvolutionPercent,
		MonthlyOutcome:           monthlyOutcome,
		NextMonthGuidance:        nextMonthGuidance,
		Insights:                 insights,
	}, nil
}

func replayOutcome(netSavings, totalSpent, previousSpent float64, growingCategory string) (string, string) {
	outcome := "insufficient_history"
	if netSavings < 0 {
		outcome = "setback"
	} else if previousSpent > 0 && totalSpent < previousSpent {
		outcome = "improvement"
	} else if previousSpent > 0 && totalSpent > previousSpent {
		outcome = "setback"
	} else if previousSpent > 0 {
		outcome = "stable"
	}

	switch {
	case netSavings < 0 && growingCategory != "":
		return outcome, fmt.Sprintf("Revise primeiro os gastos em %s e preserve saldo para os compromissos do próximo mês.", growingCategory)
	case netSavings < 0:
		return outcome, "Reduza gastos variáveis antes de assumir novos compromissos no próximo mês."
	case growingCategory != "":
		return outcome, fmt.Sprintf("Acompanhe %s no próximo mês para confirmar se o crescimento foi pontual ou recorrente.", growingCategory)
	case outcome == "improvement":
		return outcome, "Mantenha o ritmo atual e acompanhe se a redução continua no próximo mês."
	case outcome == "stable":
		return outcome, "Mantenha o acompanhamento semanal para evitar aumentos fora do ritmo atual."
	default:
		return outcome, "Continue acompanhando o ritmo semanal até existir histórico comparável suficiente."
	}
}

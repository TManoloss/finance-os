package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// GetConvenienceIndex identifica canais de conveniência e preserva a categoria
// principal/subcategoria para que Delivery não seja confundido com Alimentação.
func (s *VisualReportsService) GetConvenienceIndex(ctx context.Context, userID string) (map[string]interface{}, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.amount, COALESCE(t.merchant_name,''), t.description,
		COALESCE(c.name,''), COALESCE(parent.name,'')
		FROM transactions t JOIN connected_accounts a ON a.id=t.account_id
		LEFT JOIN categories c ON c.id=t.category_id
		LEFT JOIN categories parent ON parent.id=c.parent_id
		WHERE a.user_id=$1 AND t.direction='debit' AND t.date >= CURRENT_DATE - INTERVAL '90 days'
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := map[string]map[string]interface{}{}
	var totalSpent, totalIncome float64
	for rows.Next() {
		var amount float64
		var merchant, description, category, parent string
		if err := rows.Scan(&amount, &merchant, &description, &category, &parent); err != nil {
			return nil, err
		}
		totalSpent += amount
		text := strings.ToLower(merchant + " " + description + " " + category)
		channel := ""
		if strings.Contains(text, "ifood") || strings.Contains(text, "rappi") || strings.Contains(text, "uber eats") || strings.Contains(text, "delivery") || category == "Delivery" {
			channel = "delivery"
		}
		if channel == "" && (strings.Contains(text, "uber") || strings.Contains(text, "99app") || category == "Transporte por aplicativo") {
			channel = "transporte_por_aplicativo"
		}
		if channel == "" && (category == "Loja de conveniência" || strings.Contains(text, "conveniência") || strings.Contains(text, "conveniencia")) {
			channel = "loja_de_conveniencia"
		}
		if channel != "" {
			entry, ok := channels[channel]
			if !ok {
				entry = map[string]interface{}{"spent": 0.0, "count": 0, "category": parent}
				channels[channel] = entry
			}
			entry["spent"] = entry["spent"].(float64) + amount
			entry["count"] = entry["count"].(int) + 1
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount) FILTER (WHERE direction='credit'),0) FROM transactions t JOIN connected_accounts a ON a.id=t.account_id WHERE a.user_id=$1 AND t.date >= CURRENT_DATE - INTERVAL '90 days'`, userID).Scan(&totalIncome); err != nil {
		return nil, err
	}
	for name, entry := range channels {
		spent := entry["spent"].(float64)
		entry["name"] = name
		entry["monthly_spent"] = spent / 3
		entry["premium_assumption_percent"] = 60
		entry["estimated_premium_monthly"] = (spent - spent/1.6) / 3
	}
	convenienceTotal := 0.0
	for _, entry := range channels {
		convenienceTotal += entry["spent"].(float64)
	}
	return map[string]interface{}{"period_days": 90, "total_spent": totalSpent, "total_income": totalIncome, "convenience_spent": convenienceTotal, "monthly_convenience_spent": convenienceTotal / 3, "convenience_percentage": math.Round(convenienceTotal/math.Max(totalSpent, 1)*1000) / 10, "by_channel": channels, "quality": map[bool]string{true: "high", false: "low"}[totalSpent > 0], "note": "O prêmio é uma estimativa comparativa; a categoria e o canal vêm das transações classificadas."}, nil
}

// GetBehavioralInsights calcula padrões observáveis sem depender de um provedor de IA.
func (s *VisualReportsService) GetBehavioralInsights(ctx context.Context, userID string) (map[string]interface{}, error) {
	var txCount int
	var totalDebits, totalCredits, medianDebit float64
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount) FILTER (WHERE direction='debit'),0),
		COALESCE(SUM(amount) FILTER (WHERE direction='credit'),0),
		COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) FILTER (WHERE direction='debit'),0)
		FROM transactions t JOIN connected_accounts a ON a.id=t.account_id
		WHERE a.user_id=$1 AND t.date >= CURRENT_DATE - INTERVAL '90 days'
	`, userID).Scan(&txCount, &totalDebits, &totalCredits, &medianDebit); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT COALESCE(NULLIF(t.merchant_name,''), t.description), SUM(t.amount), COUNT(*)
		FROM transactions t JOIN connected_accounts a ON a.id=t.account_id
		WHERE a.user_id=$1 AND t.direction='debit' AND t.date >= CURRENT_DATE - INTERVAL '90 days'
		GROUP BY 1 ORDER BY SUM(t.amount) DESC LIMIT 5
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	topMerchants := make([]map[string]interface{}, 0, 5)
	for rows.Next() {
		var name string
		var amount float64
		var count int
		if err := rows.Scan(&name, &amount, &count); err != nil {
			return nil, err
		}
		topMerchants = append(topMerchants, map[string]interface{}{"name": name, "amount": amount, "count": count})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	alerts := make([]map[string]interface{}, 0)
	if medianDebit > 0 {
		var count int
		var total float64
		if err := s.db.QueryRow(ctx, `
			SELECT COUNT(*), COALESCE(SUM(t.amount),0) FROM transactions t JOIN connected_accounts a ON a.id=t.account_id
			WHERE a.user_id=$1 AND t.direction='debit' AND t.amount >= $2 * 3 AND t.date >= CURRENT_DATE - INTERVAL '90 days'
		`, userID, medianDebit).Scan(&count, &total); err != nil {
			return nil, err
		}
		if count > 0 {
			alerts = append(alerts, map[string]interface{}{"type": "high_spend", "count": count, "amount": total, "message": fmt.Sprintf("%d gasto(s) ficaram acima de três vezes a mediana dos débitos.", count)})
		}
	}
	return map[string]interface{}{
		"period":            map[string]string{"start": time.Now().AddDate(0, 0, -90).Format("2006-01-02"), "end": time.Now().Format("2006-01-02")},
		"quality":           map[bool]string{true: "high", false: "low"}[txCount >= 15],
		"confidence":        math.Min(1, float64(txCount)/30),
		"transaction_count": txCount, "total_debits": totalDebits, "total_credits": totalCredits,
		"top_merchants": topMerchants, "alerts": alerts,
	}, nil
}

// GetInvisibleSpending encontra recorrências e duplicidades diretamente no extrato.
func (s *VisualReportsService) GetInvisibleSpending(ctx context.Context, userID string) (map[string]interface{}, error) {
	rows, err := s.db.Query(ctx, `
		SELECT COALESCE(NULLIF(t.merchant_name,''), t.description), AVG(t.amount), COUNT(*),
		COUNT(DISTINCT DATE_TRUNC('month', t.date))
		FROM transactions t JOIN connected_accounts a ON a.id=t.account_id
		WHERE a.user_id=$1 AND t.direction='debit' AND t.date >= CURRENT_DATE - INTERVAL '180 days'
		GROUP BY 1 HAVING COUNT(*) >= 2 AND COUNT(DISTINCT DATE_TRUNC('month', t.date)) >= 2
		ORDER BY AVG(t.amount) DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	recurring := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name string
		var avg float64
		var count, months int
		if err := rows.Scan(&name, &avg, &count, &months); err != nil {
			return nil, err
		}
		recurring = append(recurring, map[string]interface{}{"merchant": name, "average_amount": avg, "occurrences": count, "months": months})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	dupRows, err := s.db.Query(ctx, `
		SELECT COALESCE(NULLIF(t.merchant_name,''), t.description), t.amount, COUNT(*)
		FROM transactions t JOIN connected_accounts a ON a.id=t.account_id
		WHERE a.user_id=$1 AND t.direction='debit' AND t.date >= CURRENT_DATE - INTERVAL '90 days'
		GROUP BY 1,2 HAVING COUNT(*) > 1 AND MAX(t.date) - MIN(t.date) <= 2 ORDER BY COUNT(*) DESC LIMIT 20
	`, userID)
	if err != nil {
		return nil, err
	}
	defer dupRows.Close()
	duplicates := make([]map[string]interface{}, 0)
	for dupRows.Next() {
		var name string
		var amount float64
		var count int
		if err := dupRows.Scan(&name, &amount, &count); err != nil {
			return nil, err
		}
		duplicates = append(duplicates, map[string]interface{}{"merchant": name, "amount": amount, "occurrences": count})
	}
	if err := dupRows.Err(); err != nil {
		return nil, err
	}
	return map[string]interface{}{"period_days": 180, "recurring": recurring, "duplicates": duplicates, "quality": map[bool]string{true: "high", false: "low"}[len(recurring)+len(duplicates) > 0]}, nil
}

// GetProjections cria uma projeção de fluxo simples e rastreável para três meses.
func (s *VisualReportsService) GetProjections(ctx context.Context, userID string) (map[string]interface{}, error) {
	var balance, income, expense, installments float64
	if err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(balance),0) FROM connected_accounts WHERE user_id=$1`, userID).Scan(&balance); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount) FILTER (WHERE direction='credit'),0)/3,
		COALESCE(SUM(amount) FILTER (WHERE direction='debit'),0)/3
		FROM transactions t JOIN connected_accounts a ON a.id=t.account_id
		WHERE a.user_id=$1 AND t.date >= CURRENT_DATE - INTERVAL '90 days'
	`, userID).Scan(&income, &expense); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.total_amount / NULLIF(i.installments_total,0)),0)
		FROM installments i JOIN connected_accounts a ON a.id=i.account_id
		WHERE a.user_id=$1 AND i.installment_current < i.installments_total
	`, userID).Scan(&installments); err != nil {
		return nil, err
	}
	months := buildProjectionMonths(time.Now(), balance, income, expense, installments)
	return map[string]interface{}{"period_days": 90, "starting_balance": balance, "average_monthly_income": income, "average_monthly_expense": expense, "monthly_commitments": installments, "months": months, "quality": map[bool]string{true: "high", false: "low"}[income+expense > 0]}, nil
}

func buildProjectionMonths(start time.Time, balance, income, expense, commitments float64) []map[string]interface{} {
	months := make([]map[string]interface{}, 0, 3)
	running := balance
	for i := 1; i <= 3; i++ {
		starting := running
		running += income - expense - commitments
		months = append(months, map[string]interface{}{
			"month":            start.AddDate(0, i, 0).Format("2006-01"),
			"starting_balance": starting,
			"income":           income,
			"expenses":         expense,
			"commitments":      commitments,
			"ending_balance":   running,
			"negative":         running < 0,
		})
	}
	return months
}

package service

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Subscription representa um gasto recorrente detectado.
type Subscription struct {
	MerchantName    string    `json:"merchant_name"`
	Amount          float64   `json:"amount"`
	LastDate        time.Time `json:"last_date"`
	NextEstimate    time.Time `json:"next_estimate"`
	Status          string    `json:"status"` // active, irregular
	ConfidenceScore float64   `json:"confidence_score"`
}

// SubscriptionService gerencia a identificação de assinaturas.
type SubscriptionService struct {
	db *pgxpool.Pool
}

func NewSubscriptionService(db *pgxpool.Pool) *SubscriptionService {
	return &SubscriptionService{db: db}
}

// DetectSubscriptions analisa o histórico do usuário para encontrar recorrências.
func (s *SubscriptionService) DetectSubscriptions(ctx context.Context, userID string) ([]Subscription, error) {
	// 1. Buscar transações dos últimos 90 dias agrupadas por merchant
	query := `
		SELECT 
			COALESCE(merchant_name, description) as merchant,
			amount,
			date
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date > NOW() - INTERVAL '90 days'
		ORDER BY merchant, date ASC
	`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type txInfo struct {
		Amount float64
		Date   time.Time
	}
	groups := make(map[string][]txInfo)
	for rows.Next() {
		var merchant string
		var tx txInfo
		if err := rows.Scan(&merchant, &tx.Amount, &tx.Date); err != nil {
			continue
		}
		groups[merchant] = append(groups[merchant], tx)
	}

	var subscriptions []Subscription

	// 2. Analisar cada grupo de merchant
	for merchant, txs := range groups {
		if len(txs) < 2 {
			continue // Precisa de pelo menos 2 ocorrências para detectar padrão
		}

		// Verifica periodicidade (média de dias entre transações)
		totalDays := 0
		similarAmountCount := 0
		lastAmount := txs[0].Amount

		for i := 1; i < len(txs); i++ {
			diff := txs[i].Date.Sub(txs[i-1].Date).Hours() / 24
			totalDays += int(diff)
			
			// Verifica se o valor é similar (margem de 5%)
			if math.Abs(txs[i].Amount-lastAmount) <= (lastAmount * 0.05) {
				similarAmountCount++
			}
			lastAmount = txs[i].Amount
		}

		avgDays := float64(totalDays) / float64(len(txs)-1)

		// Se a média é entre 25 e 35 dias (padrão mensal)
		if avgDays >= 25 && avgDays <= 35 && similarAmountCount >= (len(txs)-1) {
			lastTx := txs[len(txs)-1]
			nextDate := lastTx.Date.AddDate(0, 0, int(avgDays))
			
			status := "active"
			// Se a última transação foi há mais de 45 dias, a assinatura pode estar irregular/cancelada
			if time.Since(lastTx.Date).Hours()/24 > 45 {
				status = "irregular"
			}

			subscriptions = append(subscriptions, Subscription{
				MerchantName: merchant,
				Amount:       lastTx.Amount,
				LastDate:     lastTx.Date,
				NextEstimate: nextDate,
				Status:       status,
				ConfidenceScore: 0.9,
			})
		}
	}

	return subscriptions, nil
}

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ImpulseAlert struct {
	TransactionID string    `json:"transaction_id"`
	MerchantName  string    `json:"merchant_name"`
	Amount        float64   `json:"amount"`
	Signals       []string  `json:"signals"`
	Narrative     string    `json:"narrative"`
	CreatedAt     time.Time `json:"created_at"`
}

type ImpulseRadarService struct {
	db *pgxpool.Pool
}

func NewImpulseRadarService(db *pgxpool.Pool) *ImpulseRadarService {
	return &ImpulseRadarService{db: db}
}

// AnalyzeRecentTransactions analisa transações recentes em busca de impulsividade.
func (s *ImpulseRadarService) AnalyzeRecentTransactions(ctx context.Context, userID string, since time.Time) ([]ImpulseAlert, error) {
	// 1. Buscar transações das últimas 24h acima de R$50
	query := `
		SELECT t.id, COALESCE(t.merchant_name, t.description) as merchant, t.amount, t.created_at
		FROM transactions t
		JOIN connected_accounts acc ON t.account_id = acc.id
		WHERE acc.user_id = $1 
		AND t.direction = 'debit' 
		AND t.amount >= 50 
		AND t.created_at >= $2
		ORDER BY t.created_at DESC
	`
	rows, err := s.db.Query(ctx, query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []ImpulseAlert
	for rows.Next() {
		var tx struct {
			ID           string
			MerchantName string
			Amount       float64
			CreatedAt    time.Time
		}
		if err := rows.Scan(&tx.ID, &tx.MerchantName, &tx.Amount, &tx.CreatedAt); err != nil {
			continue
		}

		signals := s.detectSignals(ctx, userID, tx.ID, tx.MerchantName, tx.Amount, tx.CreatedAt)
		
		// 2+ sinais = gera alerta (conforme 13.9)
		if len(signals) >= 2 {
			narrative := s.generateNarrative(tx.MerchantName, tx.Amount, signals, tx.CreatedAt)
			alerts = append(alerts, ImpulseAlert{
				TransactionID: tx.ID,
				MerchantName:  tx.MerchantName,
				Amount:        tx.Amount,
				Signals:       signals,
				Narrative:     narrative,
				CreatedAt:     tx.CreatedAt,
			})
		}
	}

	return alerts, nil
}

func (s *ImpulseRadarService) detectSignals(ctx context.Context, userID, txID, merchant string, amount float64, createdAt time.Time) []string {
	var signals []string

	// SINAL 1: Merchant novo (primeira vez)
	var count int
	queryMerchant := `
		SELECT COUNT(*) 
		FROM transactions t 
		JOIN connected_accounts acc ON t.account_id = acc.id 
		WHERE acc.user_id = $1 AND (t.merchant_name = $2 OR t.description = $2) AND t.id != $3
	`
	err := s.db.QueryRow(ctx, queryMerchant, userID, merchant, txID).Scan(&count)
	if err == nil && count == 0 {
		signals = append(signals, "new_merchant")
	}

	// SINAL 2: Fora do horário habitual (horário de madrugada 22h-06h é um proxy simples conforme 13.9)
	hour := createdAt.Hour()
	if hour >= 22 || hour < 6 {
		signals = append(signals, "unusual_hour")
	}

	// SINAL 3: N-ésima compra acima de R$100 nos últimos 3 dias
	if amount >= 100 {
		var recentCount int
		queryFreq := `
			SELECT COUNT(*) 
			FROM transactions t 
			JOIN connected_accounts acc ON t.account_id = acc.id 
			WHERE acc.user_id = $1 AND t.amount >= 100 AND t.created_at >= $2 AND t.id != $3
		`
		err := s.db.QueryRow(ctx, queryFreq, userID, createdAt.AddDate(0, 0, -3), txID).Scan(&recentCount)
		if err == nil && recentCount >= 2 {
			signals = append(signals, "high_frequency")
		}
	}

	// SINAL 4: < 30 min após outra transação
	var quickSequence int
	querySeq := `
		SELECT COUNT(*) 
		FROM transactions t 
		JOIN connected_accounts acc ON t.account_id = acc.id 
		WHERE acc.user_id = $1 AND t.created_at BETWEEN $2 AND $3 AND t.id != $4
	`
	err = s.db.QueryRow(ctx, querySeq, userID, createdAt.Add(-30*time.Minute), createdAt.Add(30*time.Minute), txID).Scan(&quickSequence)
	if err == nil && quickSequence > 0 {
		signals = append(signals, "fast_sequence")
	}

	return signals
}

func (s *ImpulseRadarService) generateNarrative(merchant string, amount float64, signals []string, createdAt time.Time) string {
	// Exemplos do 13.9
	if contains(signals, "new_merchant") && contains(signals, "unusual_hour") {
		return fmt.Sprintf("Primeira vez nesse estabelecimento, às %02d:%02d.", createdAt.Hour(), createdAt.Minute())
	}
	if contains(signals, "high_frequency") {
		return fmt.Sprintf("Essa é sua n-ésima compra acima de R$100 em poucos dias.")
	}
	if contains(signals, "fast_sequence") {
		return "Esta compra foi realizada logo após outra transação."
	}
	
	return "Padrão de compra incomum detectado recentemente."
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

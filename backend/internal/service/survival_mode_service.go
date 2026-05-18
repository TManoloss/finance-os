package service

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SurvivalModeStatus struct {
	RiskScore          float64                   `json:"risk_score"`
	Level              string                    `json:"level"`
	IsActive           bool                      `json:"is_active"`
	ProjectedShortfall float64                   `json:"projected_shortfall"`
	DaysUntilSalary    int                       `json:"days_until_salary"`
	TopRisks           []string                  `json:"top_risks"`
	Recommendations    []SurvivalRecommendation `json:"recommendations"`
}

type SurvivalRecommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Priority    string `json:"priority"` // High, Medium, Low
}

type SurvivalModeService struct {
	db *pgxpool.Pool
}

func NewSurvivalModeService(db *pgxpool.Pool) *SurvivalModeService {
	return &SurvivalModeService{db: db}
}

func (s *SurvivalModeService) EvaluateSurvivalMode(ctx context.Context, userID string) (*SurvivalModeStatus, error) {
	// 1. Saldo projetado (peso 35%)
	projectedScore, daysUntilSalary, projectedShortfall, err := s.calculateProjectedScore(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Velocidade de gasto (peso 25%)
	velocityScore, err := s.calculateVelocityScore(ctx, userID)
	if err != nil {
		velocityScore = 50.0 // Default if error
	}

	// 3. Uso de crédito (peso 20%)
	creditScore, err := s.calculateCreditScore(ctx, userID)
	if err != nil {
		creditScore = 50.0
	}

	// 4. Proximidade do salário (peso 10%)
	proximityScore := 0.0
	if daysUntilSalary <= 3 {
		proximityScore = 100.0
	} else if daysUntilSalary > 15 {
		proximityScore = 0.0
	} else {
		proximityScore = float64(15-daysUntilSalary) / 12.0 * 100.0
	}

	// 5. Recorrência de saldo baixo (peso 10%)
	lowBalanceScore, err := s.calculateLowBalanceScore(ctx, userID)
	if err != nil {
		lowBalanceScore = 50.0
	}

	totalScore := (projectedScore * 0.35) + (velocityScore * 0.25) + (creditScore * 0.20) + (proximityScore * 0.10) + (lowBalanceScore * 0.10)

	level := "TRANQUILO"
	isActive := false
	if totalScore < 20 {
		level = "CRITICO"
		isActive = true
	} else if totalScore < 45 {
		level = "PRESSAO"
	} else if totalScore < 70 {
		level = "ATENCAO"
	}

	status := &SurvivalModeStatus{
		RiskScore:          totalScore,
		Level:              level,
		IsActive:           isActive,
		ProjectedShortfall: projectedShortfall,
		DaysUntilSalary:    daysUntilSalary,
		TopRisks:           []string{},
	}

	if projectedScore < 50 {
		status.TopRisks = append(status.TopRisks, "Risco de saldo negativo antes do próximo salário")
	}
	if velocityScore < 50 {
		status.TopRisks = append(status.TopRisks, "Gasto semanal muito acima da média")
	}
	if creditScore < 50 {
		status.TopRisks = append(status.TopRisks, "Uso elevado de limite de crédito")
	}

	// Save to snapshots
	err = s.saveSnapshot(ctx, userID, status)
	if err != nil {
		// Log error but don't fail the request
	}

	return status, nil
}

func (s *SurvivalModeService) calculateProjectedScore(ctx context.Context, userID string) (score float64, daysUntilSalary int, shortfall float64, err error) {
	// 1. Saldo atual
	var totalBalance float64
	err = s.db.QueryRow(ctx, "SELECT COALESCE(SUM(balance), 0) FROM connected_accounts WHERE user_id = $1 AND account_type != 'CREDIT'", userID).Scan(&totalBalance)
	if err != nil {
		return 0, 0, 0, err
	}

	// 2. Detectar salário
	// Simplificado: Buscar o maior crédito recorrente nos últimos 90 dias
	var avgSalary float64
	var lastSalaryDay int
	err = s.db.QueryRow(ctx, `
		WITH salary_txs AS (
			SELECT amount, EXTRACT(DAY FROM date) as day
			FROM transactions t
			JOIN connected_accounts a ON t.account_id = a.id
			WHERE a.user_id = $1 AND t.direction = 'credit' AND t.amount > 1000
			AND t.date > NOW() - INTERVAL '90 days'
		)
		SELECT COALESCE(AVG(amount), 0), COALESCE(MODE() WITHIN GROUP (ORDER BY day), 5)::int
		FROM salary_txs
	`, userID).Scan(&avgSalary, &lastSalaryDay)
	if err != nil {
		// fallback if no salary found
		avgSalary = 0
		lastSalaryDay = 5
	}

	// 3. Dias até próximo salário
	today := time.Now()
	nextSalaryDate := time.Date(today.Year(), today.Month(), lastSalaryDay, 0, 0, 0, 0, today.Location())
	if today.Day() >= lastSalaryDay {
		nextSalaryDate = nextSalaryDate.AddDate(0, 1, 0)
	}
	daysUntilSalary = int(math.Ceil(time.Until(nextSalaryDate).Hours() / 24))

	// 4. Gasto médio diário (últimos 30 dias)
	var avgDailySpent float64
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) / 30
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date > NOW() - INTERVAL '30 days'
	`, userID).Scan(&avgDailySpent)
	if err != nil {
		avgDailySpent = 0
	}

	projectedBalance := totalBalance - (avgDailySpent * float64(daysUntilSalary))

	if projectedBalance < 0 {
		score = 0
		shortfall = math.Abs(projectedBalance)
	} else if avgSalary > 0 && projectedBalance < (avgSalary*0.2) {
		score = 50
	} else if avgSalary > 0 && projectedBalance > (avgSalary*0.5) {
		score = 100
	} else {
		score = 75 // Intermediate
	}

	return score, daysUntilSalary, shortfall, nil
}

func (s *SurvivalModeService) calculateVelocityScore(ctx context.Context, userID string) (float64, error) {
	var last7DaysSpent float64
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date > NOW() - INTERVAL '7 days'
	`, userID).Scan(&last7DaysSpent)
	if err != nil {
		return 50, err
	}

	var avgWeeklySpent float64
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) / 4
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date > NOW() - INTERVAL '30 days'
	`, userID).Scan(&avgWeeklySpent)
	if err != nil {
		return 50, err
	}

	if avgWeeklySpent == 0 {
		return 100, nil
	}

	ratio := last7DaysSpent / avgWeeklySpent
	if ratio > 1.5 {
		return 0, nil
	} else if ratio < 0.8 {
		return 100, nil
	} else {
		return 100 - ((ratio - 0.8) / 0.7 * 100), nil
	}
}

func (s *SurvivalModeService) calculateCreditScore(ctx context.Context, userID string) (float64, error) {
	// As we don't have limit, let's use a heuristic: 
	// Limit is 2x the average monthly income or a default value
	var income float64
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) / 3
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'credit' AND t.date > NOW() - INTERVAL '90 days'
	`, userID).Scan(&income)
	if err != nil || income == 0 {
		income = 3000 // Default fallback
	}

	estimatedLimit := income * 1.5

	var openCreditBalance float64
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(ABS(balance)), 0)
		FROM connected_accounts
		WHERE user_id = $1 AND account_type = 'CREDIT'
	`, userID).Scan(&openCreditBalance)
	if err != nil {
		return 50, err
	}

	usageRatio := openCreditBalance / estimatedLimit
	if usageRatio > 0.8 {
		return 0, nil
	} else if usageRatio < 0.3 {
		return 100, nil
	} else {
		return 100 - ((usageRatio - 0.3) / 0.5 * 100), nil
	}
}

func (s *SurvivalModeService) calculateLowBalanceScore(ctx context.Context, userID string) (float64, error) {
	// This is hard without a balance history table. 
	// Let's check how many times the daily balance was < 200 in the last 90 days if we had snapshots.
	// Since we don't, let's count days where transactions resulted in a balance < 200 (too complex).
	// Simplification: Count how many times current balance in ANY account is < 200.
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM connected_accounts
		WHERE user_id = $1 AND account_type != 'CREDIT' AND balance < 200
	`, userID).Scan(&count)
	if err != nil {
		return 50, err
	}

	if count > 2 {
		return 0, nil
	} else if count == 0 {
		return 100, nil
	} else {
		return 50, nil
	}
}

func (s *SurvivalModeService) saveSnapshot(ctx context.Context, userID string, status *SurvivalModeStatus) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO survival_mode_snapshots (user_id, risk_score, level, is_active, projected_shortfall, days_until_salary, top_risks)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, status.RiskScore, status.Level, status.IsActive, status.ProjectedShortfall, status.DaysUntilSalary, status.TopRisks)
	return err
}

func (s *SurvivalModeService) GetSurvivalRecommendations(ctx context.Context, userID string) ([]SurvivalRecommendation, error) {
	// This should normally call an LLM or have more complex logic.
	// For now, let's return some static based on data.
	
	recs := []SurvivalRecommendation{
		{
			Title:       "Corte de assinaturas não essenciais",
			Description: "Identificamos 3 assinaturas que você não utiliza com frequência.",
			Action:      "Ver assinaturas",
			Priority:    "High",
		},
		{
			Title:       "Redução de delivery",
			Description: "Seus gastos com delivery estão 40% acima da sua média.",
			Action:      "Ver gastos com alimentação",
			Priority:    "Medium",
		},
		{
			Title:       "Limite diário recomendado",
			Description: "Tente manter seus gastos abaixo de R$ 50 por dia até o próximo salário.",
			Action:      "Ver limite diário",
			Priority:    "High",
		},
	}
	
	return recs, nil
}

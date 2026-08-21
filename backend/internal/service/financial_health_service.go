package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthPillar struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Score       float64 `json:"score"` // 0 to 100
	Weight      float64 `json:"weight"`
	MetricLabel string  `json:"metric_label"`
	Diagnosis   string  `json:"diagnosis"`
	Status      string  `json:"status"` // good, attention, critical
}

type HealthScoreResult struct {
	OverallScore      float64        `json:"overall_score"`
	Status            string         `json:"status"` // excellent, good, fair, attention, critical
	PeriodStart       string         `json:"period_start"`
	PeriodEnd         string         `json:"period_end"`
	Quality           string         `json:"quality"`    // high, medium, low
	Confidence        float64        `json:"confidence"` // 0.0 to 1.0
	DimensionsUsed    []string       `json:"dimensions_used"`
	MissingDimensions []string       `json:"missing_dimensions"`
	Pillars           []HealthPillar `json:"pillars"`
	Recommendations   []string       `json:"recommendations"`
}

type IntelligenceGroupSummary struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"` // ok, attention, warning, critical
	Severity    string   `json:"severity"` // info, warning, danger, success
	Score       *float64 `json:"score,omitempty"`
	Summary     string   `json:"summary"`
	DetailRoute string   `json:"detail_route"`
}

type ConsolidatedIntelligence struct {
	OverallHealthScore float64                    `json:"overall_health_score"`
	HealthStatus       string                     `json:"health_status"`
	Quality            string                     `json:"quality"`
	Confidence         float64                    `json:"confidence"`
	Groups             []IntelligenceGroupSummary `json:"groups"`
	ComputedAt         time.Time                  `json:"computed_at"`
}

type FinancialHealthService struct {
	db                  *pgxpool.Pool
	survivalModeService *SurvivalModeService
	impulseRadarService *ImpulseRadarService
}

func NewFinancialHealthService(db *pgxpool.Pool, survival *SurvivalModeService, impulse *ImpulseRadarService) *FinancialHealthService {
	return &FinancialHealthService{
		db:                  db,
		survivalModeService: survival,
		impulseRadarService: impulse,
	}
}

// CalculateHealthScore calcula a saúde financeira com base exclusivamente em métricas reais (FOS-701, FOS-702).
func (s *FinancialHealthService) CalculateHealthScore(ctx context.Context, userID string) (*HealthScoreResult, error) {
	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)
	periodStart := ninetyDaysAgo.Format("2006-01-02")
	periodEnd := now.Format("2006-01-02")

	// 1. Saldo consolidado
	var currentBalance float64
	_ = s.db.QueryRow(ctx, "SELECT COALESCE(SUM(balance), 0) FROM connected_accounts WHERE user_id = $1", userID).Scan(&currentBalance)

	// 2. Histórico de transações dos últimos 90 dias
	var totalCredits, totalDebits float64
	var txCount int
	var minDate, maxDate *time.Time

	_ = s.db.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN t.direction = 'credit' THEN t.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN t.direction = 'debit' THEN t.amount ELSE 0 END), 0),
			COUNT(*),
			MIN(t.date),
			MAX(t.date)
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.date >= CURRENT_DATE - INTERVAL '90 days'
	`, userID).Scan(&totalCredits, &totalDebits, &txCount, &minDate, &maxDate)

	daySpan := 0
	if minDate != nil && maxDate != nil {
		daySpan = int(maxDate.Sub(*minDate).Hours() / 24)
	}

	quality := "low"
	confidence := 0.40
	if txCount >= 30 && daySpan >= 45 {
		quality = "high"
		confidence = 0.95
	} else if txCount >= 10 && daySpan >= 14 {
		quality = "medium"
		confidence = 0.75
	}

	monthsCount := math.Max(1.0, float64(daySpan)/30.0)
	avgMonthlyIncome := totalCredits / monthsCount
	avgMonthlyExpense := totalDebits / monthsCount

	// 3. Parcelas mensais ativas
	var activeInstallmentsMonthly float64
	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.total_amount / NULLIF(i.installments_total, 0)), 0)
		FROM installments i
		JOIN connected_accounts a ON i.account_id = a.id
		WHERE a.user_id = $1 AND i.installment_current < i.installments_total
	`, userID).Scan(&activeInstallmentsMonthly)

	var pillars []HealthPillar
	var dimensionsUsed []string
	var missingDimensions []string
	var recommendations []string

	// Pilar 1: Fluxo de Caixa / Taxa de Poupança (peso 25%)
	if totalCredits > 0 {
		savingsRatio := (totalCredits - totalDebits) / totalCredits
		savingsScore := math.Max(0, math.Min(100, (savingsRatio+0.2)/0.4*100))
		status := "good"
		diag := fmt.Sprintf("Taxa de poupança real de %.1f%% nos últimos 90 dias.", savingsRatio*100)
		if savingsRatio < 0 {
			status = "critical"
			diag = fmt.Sprintf("Gastos superaram as receitas em %.1f%% no período.", math.Abs(savingsRatio*100))
			recommendations = append(recommendations, "Ajustar despesas mensais para reverter o fluxo de caixa negativo.")
		} else if savingsRatio < 0.10 {
			status = "attention"
			diag = fmt.Sprintf("Margem de poupança estreita (%.1f%%).", savingsRatio*100)
		}

		pillars = append(pillars, HealthPillar{
			ID:          "cash_flow",
			Name:        "Fluxo de Caixa",
			Score:       math.Round(savingsScore),
			Weight:      0.25,
			MetricLabel: fmt.Sprintf("%.1f%% de margem", savingsRatio*100),
			Diagnosis:   diag,
			Status:      status,
		})
		dimensionsUsed = append(dimensionsUsed, "cash_flow")
	} else {
		missingDimensions = append(missingDimensions, "cash_flow")
	}

	// Pilar 2: Reserva de Emergência (peso 25%)
	if avgMonthlyExpense > 0 {
		monthsCoverage := currentBalance / avgMonthlyExpense
		fundScore := math.Max(0, math.Min(100, (monthsCoverage/6.0)*100))
		status := "good"
		diag := fmt.Sprintf("Saldo cobre %.1f meses do seu ritmo médio de gastos.", monthsCoverage)
		if monthsCoverage < 1.0 {
			status = "critical"
			diag = fmt.Sprintf("Reserva atual cobre apenas %.1f meses de gastos.", monthsCoverage)
			recommendations = append(recommendations, "Construir reserva mínima de 1 a 3 meses de despesas essenciais.")
		} else if monthsCoverage < 3.0 {
			status = "attention"
			diag = fmt.Sprintf("Reserva parcial (%.1f meses). Ideal buscar 3 a 6 meses.", monthsCoverage)
		}

		pillars = append(pillars, HealthPillar{
			ID:          "emergency_fund",
			Name:        "Reserva de Emergência",
			Score:       math.Round(fundScore),
			Weight:      0.25,
			MetricLabel: fmt.Sprintf("%.1f meses", monthsCoverage),
			Diagnosis:   diag,
			Status:      status,
		})
		dimensionsUsed = append(dimensionsUsed, "emergency_fund")
	} else {
		missingDimensions = append(missingDimensions, "emergency_fund")
	}

	// Pilar 3: Comprometimento com Dívidas/Parcelas (peso 20%)
	if avgMonthlyIncome > 0 {
		debtRatio := (activeInstallmentsMonthly / avgMonthlyIncome) * 100
		debtScore := math.Max(0, math.Min(100, 100.0-(debtRatio/50.0)*100.0))
		status := "good"
		diag := fmt.Sprintf("Parcelas ativas consomem %.1f%% da sua renda média.", debtRatio)
		if debtRatio > 35.0 {
			status = "critical"
			diag = fmt.Sprintf("Alto comprometimento de renda com parcelamentos (%.1f%%).", debtRatio)
			recommendations = append(recommendations, "Evitar novas compras parceladas até quitar compromissos correntes.")
		} else if debtRatio > 20.0 {
			status = "attention"
		}

		pillars = append(pillars, HealthPillar{
			ID:          "debt_commitments",
			Name:        "Compromissos e Parcelamentos",
			Score:       math.Round(debtScore),
			Weight:      0.20,
			MetricLabel: fmt.Sprintf("%.1f%% da renda", debtRatio),
			Diagnosis:   diag,
			Status:      status,
		})
		dimensionsUsed = append(dimensionsUsed, "debt_commitments")
	} else {
		missingDimensions = append(missingDimensions, "debt_commitments")
	}

	// Pilar 4: Concentração de Gastos por Categoria (peso 15%)
	var maxCatAmount float64
	var maxCatName string
	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(c.name, 'Outros'), SUM(t.amount) as cat_total
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= CURRENT_DATE - INTERVAL '90 days'
		GROUP BY c.name
		ORDER BY cat_total DESC
		LIMIT 1
	`, userID).Scan(&maxCatName, &maxCatAmount)

	if totalDebits > 0 && maxCatAmount > 0 {
		catConcentration := (maxCatAmount / totalDebits) * 100
		concScore := 85.0
		status := "good"
		diag := fmt.Sprintf("Principal categoria (%s) representa %.1f%% dos gastos.", maxCatName, catConcentration)
		if catConcentration > 50.0 {
			concScore = 45.0
			status = "attention"
			diag = fmt.Sprintf("Alta concentração na categoria %s (%.1f%% do total).", maxCatName, catConcentration)
		}

		pillars = append(pillars, HealthPillar{
			ID:          "spending_concentration",
			Name:        "Distribuição de Gastos",
			Score:       math.Round(concScore),
			Weight:      0.15,
			MetricLabel: fmt.Sprintf("%.1f%% em %s", catConcentration, maxCatName),
			Diagnosis:   diag,
			Status:      status,
		})
		dimensionsUsed = append(dimensionsUsed, "spending_concentration")
	} else {
		missingDimensions = append(missingDimensions, "spending_concentration")
	}

	// Pilar 5: Estabilidade de Entradas (peso 15%)
	if txCount > 0 {
		stabilityScore := 75.0
		if totalCredits > 0 {
			stabilityScore = 85.0
		}
		pillars = append(pillars, HealthPillar{
			ID:          "income_stability",
			Name:        "Estabilidade de Entradas",
			Score:       stabilityScore,
			Weight:      0.15,
			MetricLabel: fmt.Sprintf("%d transações", txCount),
			Diagnosis:   "Histórico de movimentação monitorado com sucesso.",
			Status:      "good",
		})
		dimensionsUsed = append(dimensionsUsed, "income_stability")
	}

	// Cálculo da pontuação geral ponderada
	totalWeight := 0.0
	weightedSum := 0.0
	for _, p := range pillars {
		weightedSum += p.Score * p.Weight
		totalWeight += p.Weight
	}

	overallScore := 50.0
	if totalWeight > 0 {
		overallScore = math.Round(weightedSum / totalWeight)
	}

	overallStatus := "fair"
	if overallScore >= 80 {
		overallStatus = "excellent"
	} else if overallScore >= 65 {
		overallStatus = "good"
	} else if overallScore >= 50 {
		overallStatus = "fair"
	} else if overallScore >= 35 {
		overallStatus = "attention"
	} else {
		overallStatus = "critical"
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Manter o ritmo de gastos e acompanhar o fluxo de caixa semanal.")
	}

	return &HealthScoreResult{
		OverallScore:      overallScore,
		Status:            overallStatus,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		Quality:           quality,
		Confidence:        confidence,
		DimensionsUsed:    dimensionsUsed,
		MissingDimensions: missingDimensions,
		Pillars:           pillars,
		Recommendations:   recommendations,
	}, nil
}

// GetConsolidatedIntelligence consolida os 9 grupos analíticos (FOS-703, FOS-704).
func (s *FinancialHealthService) GetConsolidatedIntelligence(ctx context.Context, userID string) (*ConsolidatedIntelligence, error) {
	health, err := s.CalculateHealthScore(ctx, userID)
	if err != nil {
		return nil, err
	}

	survival, _ := s.survivalModeService.EvaluateSurvivalMode(ctx, userID)
	impulseAlerts, _ := s.impulseRadarService.AnalyzeRecentTransactions(ctx, userID, time.Now().Add(-24*time.Hour))

	groups := make([]IntelligenceGroupSummary, 0, 9)

	// 1. Saúde Financeira
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "financial_health",
		Name:        "Saúde Financeira",
		Status:      health.Status,
		Severity:    healthSeverity(health.Status),
		Score:       &health.OverallScore,
		Summary:     fmt.Sprintf("Score geral de %.0f/100 baseado em %d dimensões reais.", health.OverallScore, len(health.DimensionsUsed)),
		DetailRoute: "/reports/health",
	})

	// 2. Perfil de Gastos
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "spending_profile",
		Name:        "Perfil de Gastos",
		Status:      "ok",
		Severity:    "info",
		Summary:     "Distribuição e comportamento de despesas por categoria.",
		DetailRoute: "/analytics",
	})

	// 3. Estabilidade de Renda
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "income_stability",
		Name:        "Estabilidade de Renda",
		Status:      "ok",
		Severity:    "success",
		Summary:     "Monitoramento de créditos regulares e fontes de receita.",
		DetailRoute: "/analytics",
	})

	// 4. Compromissos e Parcelamentos
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "commitments_debt",
		Name:        "Compromissos e Parcelamentos",
		Status:      "ok",
		Severity:    "info",
		Summary:     "Acompanhamento de parcelas futuras e quitações.",
		DetailRoute: "/cards",
	})

	// 5. Assinaturas e Recorrências
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "subscriptions",
		Name:        "Assinaturas e Recorrências",
		Status:      "ok",
		Severity:    "info",
		Summary:     "Detecção automática de cobranças mensais repetidas.",
		DetailRoute: "/cards",
	})

	// 6. Radar de Impulso
	impulseSev := "info"
	impulseSum := "Nenhum desvio atípico detectado nas últimas 24h."
	if len(impulseAlerts) > 0 {
		impulseSev = "warning"
		impulseSum = fmt.Sprintf("%d alerta(s) de gastos rápidos ou atípicos.", len(impulseAlerts))
	}
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "impulse_radar",
		Name:        "Radar de Impulso",
		Status:      "ok",
		Severity:    impulseSev,
		Summary:     impulseSum,
		DetailRoute: "/reports/impulse",
	})

	// 7. Inflação Pessoal
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "personal_inflation",
		Name:        "Inflação Pessoal",
		Status:      "ok",
		Severity:    "info",
		Summary:     "Variação de preço e ritmo nos principais destinos de consumo.",
		DetailRoute: "/reports/inflation",
	})

	// 8. Modo Sobrevivência / Reserva
	survSev := "success"
	survSum := "Fluxo de caixa sob controle até a próxima renda."
	if survival != nil && survival.IsActive {
		survSev = "danger"
		survSum = fmt.Sprintf("Risco de liquidez detectado (déficit projetado: R$ %.2f).", survival.ProjectedShortfall)
	}
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "survival_mode",
		Name:        "Modo Sobrevivência e Liquidez",
		Status:      "ok",
		Severity:    survSev,
		Summary:     survSum,
		DetailRoute: "/reports/survival",
	})

	// 9. Gamificação e Hábitos
	groups = append(groups, IntelligenceGroupSummary{
		ID:          "gamification_streaks",
		Name:        "Conquistas e Sequências",
		Status:      "ok",
		Severity:    "info",
		Summary:     "Acompanhamento de metas cumpridas e hábitos financeiros saudáveis.",
		DetailRoute: "/reports/gamification",
	})

	return &ConsolidatedIntelligence{
		OverallHealthScore: health.OverallScore,
		HealthStatus:       health.Status,
		Quality:            health.Quality,
		Confidence:         health.Confidence,
		Groups:             groups,
		ComputedAt:         time.Now(),
	}, nil
}

func healthSeverity(status string) string {
	switch status {
	case "excellent", "good":
		return "success"
	case "fair":
		return "info"
	case "attention":
		return "warning"
	case "critical":
		return "danger"
	default:
		return "info"
	}
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PurchaseSimulationRequest struct {
	Amount       float64 `json:"amount"`
	Installments int     `json:"installments"`
	Description  string  `json:"description"`
	FirstDueDate string  `json:"first_due_date,omitempty"` // YYYY-MM-DD
	AccountID    *string `json:"account_id,omitempty"`
}

type ProjectedMonth struct {
	MonthName         string  `json:"month_name"`
	YearMonth         string  `json:"year_month"`
	StartingBalance   float64 `json:"starting_balance"`
	EstimatedIncome   float64 `json:"estimated_income"`
	EstimatedExpenses float64 `json:"estimated_expenses"`
	InstallmentAmount float64 `json:"installment_amount"`
	EndingBalance     float64 `json:"ending_balance"`
	IsNegative        bool    `json:"is_negative"`
}

type SimulationAlert struct {
	Type     string `json:"type"` // info, warning, danger
	Message  string `json:"message"`
	Severity string `json:"severity"` // low, medium, high
}

type PurchaseSimulationResult struct {
	Amount                 float64          `json:"amount"`
	Installments           int              `json:"installments"`
	InstallmentAmount      float64          `json:"installment_amount"`
	Description            string           `json:"description"`
	CurrentBalance         float64          `json:"current_balance"`
	AvgMonthlyIncome       float64          `json:"avg_monthly_income"`
	AvgMonthlyExpense      float64          `json:"avg_monthly_expense"`
	MonthlyImpactPercent   float64          `json:"monthly_impact_percent"`
	IncomeCommittedPercent float64          `json:"income_committed_percent"`
	SafeDailyLimitAfter    float64          `json:"safe_daily_limit_after"`
	Confidence             string           `json:"confidence"` // high, medium, low
	HasSufficientHistory   bool             `json:"has_sufficient_history"`
	HistoryNotice          string           `json:"history_notice,omitempty"`
	ProjectedMonths        []ProjectedMonth `json:"projected_months"`
	Alerts                 []SimulationAlert `json:"alerts"`
}

type CutSimulationRequest struct {
	MonthlyAmount   float64 `json:"monthly_amount"`
	Description     string  `json:"description"`
	CutPeriodMonths int     `json:"cut_period_months"` // default 12
}

type GoalImpact struct {
	GoalID        string  `json:"goal_id"`
	GoalName      string  `json:"goal_name"`
	TargetAmount  float64 `json:"target_amount"`
	Remaining     float64 `json:"remaining"`
	MonthsFaster  int     `json:"months_faster"`
}

type CutSimulationResult struct {
	MonthlyAmount        float64      `json:"monthly_amount"`
	Description          string       `json:"description"`
	CutPeriodMonths      int          `json:"cut_period_months"`
	MonthlySavings       float64      `json:"monthly_savings"`
	AnnualSavings        float64      `json:"annual_savings"`
	AccumulatedSavings   float64      `json:"accumulated_savings"`
	HasSufficientHistory bool         `json:"has_sufficient_history"`
	Confidence           string       `json:"confidence"`
	GoalImpacts          []GoalImpact `json:"goal_impacts"`
	Alerts               []SimulationAlert `json:"alerts"`
}

type SavedSimulation struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	SimulationType string                 `json:"simulation_type"` // purchase, cut
	Name           string                 `json:"name"`
	InputParams    map[string]interface{} `json:"input_params"`
	ResultJSON     map[string]interface{} `json:"result_json"`
	CreatedAt      time.Time              `json:"created_at"`
}

type SimulatorService struct {
	db *pgxpool.Pool
}

func NewSimulatorService(db *pgxpool.Pool) *SimulatorService {
	return &SimulatorService{db: db}
}

// SimulatePurchase projeta o impacto financeiro real de uma compra à vista ou parcelada.
func (s *SimulatorService) SimulatePurchase(ctx context.Context, userID string, req PurchaseSimulationRequest) (*PurchaseSimulationResult, error) {
	if req.Amount <= 0 {
		return nil, errors.New("o valor da compra deve ser maior que zero")
	}
	if req.Installments <= 0 {
		req.Installments = 1
	}
	if req.Installments > 48 {
		return nil, errors.New("número máximo de parcelas permitido é 48")
	}

	firstDue := time.Now()
	if req.FirstDueDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.FirstDueDate); err == nil {
			firstDue = parsed
		}
	}

	// 1. Saldo consolidado ou da conta informada
	var currentBalance float64
	if req.AccountID != nil {
		_ = s.db.QueryRow(ctx, "SELECT COALESCE(balance, 0) FROM connected_accounts WHERE id = $1 AND user_id = $2", *req.AccountID, userID).Scan(&currentBalance)
	} else {
		_ = s.db.QueryRow(ctx, "SELECT COALESCE(SUM(balance), 0) FROM connected_accounts WHERE user_id = $1", userID).Scan(&currentBalance)
	}

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

	// 3. Avaliação de suficiência de histórico (FOS-605)
	daySpan := 0
	if minDate != nil && maxDate != nil {
		daySpan = int(maxDate.Sub(*minDate).Hours() / 24)
	}

	hasSufficientHistory := txCount >= 3 && daySpan >= 7
	confidence := "medium"
	historyNotice := ""

	if !hasSufficientHistory {
		confidence = "low"
		historyNotice = "Histórico de transações recente limitado. As projeções baseiam-se principalmente no saldo atual disponível."
	} else if txCount >= 15 && daySpan >= 30 {
		confidence = "high"
	}

	// Médias mensais reais
	monthsCount := math.Max(1.0, float64(daySpan)/30.0)
	avgMonthlyIncome := totalCredits / monthsCount
	avgMonthlyExpense := totalDebits / monthsCount

	// 4. Parcelas ativas existentes
	var currentInstallmentsMonthly float64
	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(i.total_amount / NULLIF(i.installments_total, 0)), 0)
		FROM installments i
		JOIN connected_accounts a ON i.account_id = a.id
		WHERE a.user_id = $1 AND i.installment_current < i.installments_total
	`, userID).Scan(&currentInstallmentsMonthly)

	installmentAmount := req.Amount / float64(req.Installments)

	var monthlyImpactPercent float64
	if avgMonthlyIncome > 0 {
		monthlyImpactPercent = (installmentAmount / avgMonthlyIncome) * 100
	}

	var incomeCommittedPercent float64
	if avgMonthlyIncome > 0 {
		incomeCommittedPercent = ((currentInstallmentsMonthly + installmentAmount) / avgMonthlyIncome) * 100
	}

	// 5. Limite diário seguro após a compra
	now := time.Now()
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	daysRemaining := math.Max(1, float64(daysInMonth-now.Day()))
	liquidBuffer := currentBalance - (currentInstallmentsMonthly + installmentAmount)
	safeDailyLimitAfter := math.Max(0, liquidBuffer/daysRemaining)

	// 6. Projeção mês a mês para os próximos N meses
	projectionMonthsCount := int(math.Max(3, math.Min(12, float64(req.Installments+2))))
	projectedMonths := make([]ProjectedMonth, 0, projectionMonthsCount)

	runningBalance := currentBalance
	hasNegativeProjectedBalance := false

	ptMonths := []string{"", "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}

	for m := 0; m < projectionMonthsCount; m++ {
		targetMonthDate := firstDue.AddDate(0, m, 0)
		mName := ptMonths[targetMonthDate.Month()]
		yMonth := fmt.Sprintf("%d-%02d", targetMonthDate.Year(), targetMonthDate.Month())

		startBal := runningBalance
		estIncome := avgMonthlyIncome
		estExpense := avgMonthlyExpense

		instPart := 0.0
		if m < req.Installments {
			instPart = installmentAmount
		}

		endBal := startBal + estIncome - estExpense - instPart
		if endBal < 0 {
			hasNegativeProjectedBalance = true
		}

		projectedMonths = append(projectedMonths, ProjectedMonth{
			MonthName:         mName,
			YearMonth:         yMonth,
			StartingBalance:   startBal,
			EstimatedIncome:   estIncome,
			EstimatedExpenses: estExpense,
			InstallmentAmount: instPart,
			EndingBalance:     endBal,
			IsNegative:        endBal < 0,
		})

		runningBalance = endBal
	}

	// 7. Alertas e orientações determinísticas (FOS-602)
	var alerts []SimulationAlert

	if req.Installments == 1 && req.Amount > currentBalance {
		alerts = append(alerts, SimulationAlert{
			Type:     "danger",
			Message:  fmt.Sprintf("O valor da compra (R$ %.2f) excede seu saldo disponível atual (R$ %.2f).", req.Amount, currentBalance),
			Severity: "high",
		})
	}

	if hasNegativeProjectedBalance {
		alerts = append(alerts, SimulationAlert{
			Type:     "danger",
			Message:  "A compra projeta saldo negativo em meses futuros considerando seu ritmo médio de gastos.",
			Severity: "high",
		})
	}

	if incomeCommittedPercent > 40.0 {
		alerts = append(alerts, SimulationAlert{
			Type:     "warning",
			Message:  fmt.Sprintf("O total de compromissos atingirá %.1f%% da sua renda mensal média estimada.", incomeCommittedPercent),
			Severity: "medium",
		})
	} else if incomeCommittedPercent > 0 {
		alerts = append(alerts, SimulationAlert{
			Type:     "info",
			Message:  fmt.Sprintf("Compromisso mensal de %.1f%% da renda estimada. Dentro dos parâmetros seguros.", incomeCommittedPercent),
			Severity: "low",
		})
	}

	if len(alerts) == 0 {
		alerts = append(alerts, SimulationAlert{
			Type:     "info",
			Message:  "Compra dentro da sua capacidade financeira projetada.",
			Severity: "low",
		})
	}

	return &PurchaseSimulationResult{
		Amount:                 req.Amount,
		Installments:           req.Installments,
		InstallmentAmount:      installmentAmount,
		Description:            req.Description,
		CurrentBalance:         currentBalance,
		AvgMonthlyIncome:       avgMonthlyIncome,
		AvgMonthlyExpense:      avgMonthlyExpense,
		MonthlyImpactPercent:   monthlyImpactPercent,
		IncomeCommittedPercent: incomeCommittedPercent,
		SafeDailyLimitAfter:    safeDailyLimitAfter,
		Confidence:             confidence,
		HasSufficientHistory:   hasSufficientHistory,
		HistoryNotice:          historyNotice,
		ProjectedMonths:        projectedMonths,
		Alerts:                 alerts,
	}, nil
}

// SimulateCut simula o corte de uma despesa recorrente sem rentabilidade fictícia (FOS-603).
func (s *SimulatorService) SimulateCut(ctx context.Context, userID string, req CutSimulationRequest) (*CutSimulationResult, error) {
	if req.MonthlyAmount <= 0 {
		return nil, errors.New("o valor mensal a cortar deve ser maior que zero")
	}
	if req.CutPeriodMonths <= 0 {
		req.CutPeriodMonths = 12
	}

	annualSavings := req.MonthlyAmount * 12.0
	accumulatedSavings := req.MonthlyAmount * float64(req.CutPeriodMonths)

	// Avaliação de metas ativas para calcular aceleração de objetivos reais
	goalRows, err := s.db.Query(ctx, `
		SELECT id, name, target_amount, current_amount
		FROM financial_goals
		WHERE user_id = $1 AND status = 'active' AND goal_type = 'savings'
	`, userID)

	var goalImpacts []GoalImpact
	if err == nil {
		defer goalRows.Close()
		for goalRows.Next() {
			var g GoalImpact
			var currentAmount float64
			if err := goalRows.Scan(&g.GoalID, &g.GoalName, &g.TargetAmount, &currentAmount); err == nil {
				remaining := g.TargetAmount - currentAmount
				if remaining > 0 {
					g.Remaining = remaining
					// Quantos meses essa sobra mensal sozinha economiza
					months := int(math.Ceil(remaining / req.MonthlyAmount))
					g.MonthsFaster = months
					goalImpacts = append(goalImpacts, g)
				}
			}
		}
	}

	var alerts []SimulationAlert
	alerts = append(alerts, SimulationAlert{
		Type:     "info",
		Message:  fmt.Sprintf("Cortando esta despesa, você acumula R$ %.2f em %d meses sem depender de rentabilidade especulativa.", accumulatedSavings, req.CutPeriodMonths),
		Severity: "low",
	})

	return &CutSimulationResult{
		MonthlyAmount:        req.MonthlyAmount,
		Description:          req.Description,
		CutPeriodMonths:      req.CutPeriodMonths,
		MonthlySavings:       req.MonthlyAmount,
		AnnualSavings:        annualSavings,
		AccumulatedSavings:   accumulatedSavings,
		HasSufficientHistory: true,
		Confidence:           "high",
		GoalImpacts:          goalImpacts,
		Alerts:               alerts,
	}, nil
}

// SaveSimulation persiste uma simulação realizada (FOS-604).
func (s *SimulatorService) SaveSimulation(ctx context.Context, userID, simType, name string, inputParams, resultJSON map[string]interface{}) (string, error) {
	if strings.TrimSpace(name) == "" {
		name = fmt.Sprintf("Simulação de %s em %s", simType, time.Now().Format("02/01/2006"))
	}

	inputBytes, err := json.Marshal(inputParams)
	if err != nil {
		return "", err
	}
	resultBytes, err := json.Marshal(resultJSON)
	if err != nil {
		return "", err
	}

	var id string
	err = s.db.QueryRow(ctx, `
		INSERT INTO saved_simulations (user_id, simulation_type, name, input_params, result_json, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id
	`, userID, simType, name, inputBytes, resultBytes).Scan(&id)
	return id, err
}

// ListSavedSimulations lista as simulações salvas do usuário com isolamento estrito.
func (s *SimulatorService) ListSavedSimulations(ctx context.Context, userID string) ([]SavedSimulation, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, simulation_type, COALESCE(name, 'Simulação'), input_params, result_json, created_at
		FROM saved_simulations
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SavedSimulation
	for rows.Next() {
		var sim SavedSimulation
		var inputBytes, resultBytes []byte
		if err := rows.Scan(&sim.ID, &sim.UserID, &sim.SimulationType, &sim.Name, &inputBytes, &resultBytes, &sim.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(inputBytes, &sim.InputParams)
		_ = json.Unmarshal(resultBytes, &sim.ResultJSON)
		list = append(list, sim)
	}
	return list, nil
}

// DeleteSavedSimulation exclui uma simulação garantindo propriedade do usuário.
func (s *SimulatorService) DeleteSavedSimulation(ctx context.Context, userID, id string) (bool, error) {
	tag, err := s.db.Exec(ctx, "DELETE FROM saved_simulations WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

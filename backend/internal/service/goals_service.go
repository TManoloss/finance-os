package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GoalType string

const (
	GoalSavings       GoalType = "savings"
	GoalDebtPayoff    GoalType = "debt_payoff"
	GoalSpendingLimit GoalType = "spending_limit"
	GoalIncomeTarget  GoalType = "income_target"
)

type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusPaused    GoalStatus = "paused"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusFailed    GoalStatus = "failed"
)

type GoalAdjustment struct {
	ID        string    `json:"id"`
	GoalID    string    `json:"goal_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Note      string    `json:"note,omitempty"`
	Date      time.Time `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

type FinancialGoal struct {
	ID                  string           `json:"id"`
	UserID              string           `json:"user_id"`
	Name                string           `json:"name"`
	GoalType            GoalType         `json:"goal_type"`
	TargetAmount        float64          `json:"target_amount"`
	CurrentAmount       float64          `json:"current_amount"`
	InitialAmount       float64          `json:"initial_amount"`
	StartDate           time.Time        `json:"start_date"`
	TargetDate          *time.Time       `json:"target_date,omitempty"`
	CategoryID          *string          `json:"category_id,omitempty"`
	CategoryName        *string          `json:"category_name,omitempty"`
	AccountID           *string          `json:"account_id,omitempty"`
	AccountName         *string          `json:"account_name,omitempty"`
	InstallmentID       *string          `json:"installment_id,omitempty"`
	InstallmentMerchant *string          `json:"installment_merchant,omitempty"`
	Status              string           `json:"status"` // active, paused, completed, failed
	Adjustments         []GoalAdjustment `json:"adjustments,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type PlanningTimelineItem struct {
	Type      string                 `json:"type"` // installment, subscription, recurring_income, goal_deadline
	Title     string                 `json:"title"`
	Amount    float64                `json:"amount"`
	Direction string                 `json:"direction"` // debit, credit, info
	Date      time.Time              `json:"date"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type GoalsService struct {
	db *pgxpool.Pool
}

func NewGoalsService(db *pgxpool.Pool) *GoalsService {
	return &GoalsService{db: db}
}

// ListGoals retorna todas as metas do usuário após recalcular o progresso.
func (s *GoalsService) ListGoals(ctx context.Context, userID string) ([]FinancialGoal, error) {
	_ = s.UpdateGoalProgress(ctx, userID)

	query := `
		SELECT 
			g.id, g.user_id, g.name, g.goal_type, g.target_amount, g.current_amount, g.initial_amount,
			g.start_date, g.target_date, g.category_id, c.name, g.account_id, a.account_name,
			g.installment_id, i.merchant_name, g.status, g.created_at, g.updated_at
		FROM financial_goals g
		LEFT JOIN categories c ON g.category_id = c.id
		LEFT JOIN connected_accounts a ON g.account_id = a.id
		LEFT JOIN installments i ON g.installment_id = i.id
		WHERE g.user_id = $1
		ORDER BY g.created_at DESC
	`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []FinancialGoal
	for rows.Next() {
		var g FinancialGoal
		var catName, accName, instMerchant *string
		err := rows.Scan(
			&g.ID, &g.UserID, &g.Name, &g.GoalType, &g.TargetAmount, &g.CurrentAmount, &g.InitialAmount,
			&g.StartDate, &g.TargetDate, &g.CategoryID, &catName, &g.AccountID, &accName,
			&g.InstallmentID, &instMerchant, &g.Status, &g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		g.CategoryName = catName
		g.AccountName = accName
		g.InstallmentMerchant = instMerchant
		goals = append(goals, g)
	}
	return goals, nil
}

// GetGoal retorna uma meta específica com seus ajustes manuais.
func (s *GoalsService) GetGoal(ctx context.Context, userID, goalID string) (*FinancialGoal, error) {
	query := `
		SELECT 
			g.id, g.user_id, g.name, g.goal_type, g.target_amount, g.current_amount, g.initial_amount,
			g.start_date, g.target_date, g.category_id, c.name, g.account_id, a.account_name,
			g.installment_id, i.merchant_name, g.status, g.created_at, g.updated_at
		FROM financial_goals g
		LEFT JOIN categories c ON g.category_id = c.id
		LEFT JOIN connected_accounts a ON g.account_id = a.id
		LEFT JOIN installments i ON g.installment_id = i.id
		WHERE g.id = $1 AND g.user_id = $2
	`
	var g FinancialGoal
	var catName, accName, instMerchant *string
	err := s.db.QueryRow(ctx, query, goalID, userID).Scan(
		&g.ID, &g.UserID, &g.Name, &g.GoalType, &g.TargetAmount, &g.CurrentAmount, &g.InitialAmount,
		&g.StartDate, &g.TargetDate, &g.CategoryID, &catName, &g.AccountID, &accName,
		&g.InstallmentID, &instMerchant, &g.Status, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	g.CategoryName = catName
	g.AccountName = accName
	g.InstallmentMerchant = instMerchant

	adjustments, err := s.ListAdjustments(ctx, userID, goalID)
	if err == nil {
		g.Adjustments = adjustments
	}

	return &g, nil
}

// CreateGoal cria uma nova meta e calcula seu progresso inicial.
func (s *GoalsService) CreateGoal(ctx context.Context, g FinancialGoal) (string, error) {
	if strings.TrimSpace(g.Name) == "" {
		return "", errors.New("nome da meta é obrigatório")
	}
	if g.TargetAmount <= 0 {
		return "", errors.New("valor alvo deve ser maior que zero")
	}
	if g.StartDate.IsZero() {
		g.StartDate = time.Now()
	}
	if g.Status == "" {
		g.Status = string(GoalStatusActive)
	}

	// Se for economia (savings) e initial_amount não foi informado, busca o saldo atual da conta
	if g.GoalType == GoalSavings && g.InitialAmount == 0 && g.AccountID != nil {
		var currentBalance float64
		_ = s.db.QueryRow(ctx, "SELECT COALESCE(balance, 0) FROM connected_accounts WHERE id = $1 AND user_id = $2", *g.AccountID, g.UserID).Scan(&currentBalance)
		g.InitialAmount = currentBalance
	}

	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO financial_goals (
			user_id, name, goal_type, target_amount, current_amount, initial_amount,
			start_date, target_date, category_id, account_id, installment_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		RETURNING id
	`, g.UserID, g.Name, g.GoalType, g.TargetAmount, 0, g.InitialAmount, g.StartDate, g.TargetDate, g.CategoryID, g.AccountID, g.InstallmentID, g.Status).Scan(&id)

	if err != nil {
		return "", err
	}

	// Recalcula progresso
	_ = s.UpdateGoalProgress(ctx, g.UserID)
	return id, nil
}

// UpdateGoal atualiza campos de uma meta com isolamento por usuário.
func (s *GoalsService) UpdateGoal(ctx context.Context, userID, goalID string, updates map[string]interface{}) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM financial_goals WHERE id = $1 AND user_id = $2)", goalID, userID).Scan(&exists)
	if err != nil || !exists {
		return false, err
	}

	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{goalID, userID}
	argIdx := 3

	if name, ok := updates["name"].(string); ok && strings.TrimSpace(name) != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, strings.TrimSpace(name))
		argIdx++
	}
	if target, ok := updates["target_amount"].(float64); ok && target > 0 {
		setClauses = append(setClauses, fmt.Sprintf("target_amount = $%d", argIdx))
		args = append(args, target)
		argIdx++
	}
	if status, ok := updates["status"].(string); ok && status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if targetDateStr, ok := updates["target_date"].(string); ok {
		if targetDateStr == "" {
			setClauses = append(setClauses, "target_date = NULL")
		} else if parsed, err := time.Parse("2006-01-02", targetDateStr); err == nil {
			setClauses = append(setClauses, fmt.Sprintf("target_date = $%d", argIdx))
			args = append(args, parsed)
			argIdx++
		}
	}

	query := fmt.Sprintf("UPDATE financial_goals SET %s WHERE id = $1 AND user_id = $2", strings.Join(setClauses, ", "))
	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return false, err
	}

	_ = s.UpdateGoalProgress(ctx, userID)
	return true, nil
}

// DeleteGoal exclui uma meta garantindo propriedade.
func (s *GoalsService) DeleteGoal(ctx context.Context, userID, goalID string) (bool, error) {
	tag, err := s.db.Exec(ctx, "DELETE FROM financial_goals WHERE id = $1 AND user_id = $2", goalID, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// AddAdjustment adiciona um ajuste manual a uma meta e recalcula seu progresso.
func (s *GoalsService) AddAdjustment(ctx context.Context, adj GoalAdjustment) (string, error) {
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM financial_goals WHERE id = $1 AND user_id = $2)", adj.GoalID, adj.UserID).Scan(&exists)
	if err != nil || !exists {
		return "", errors.New("meta não encontrada para este usuário")
	}

	if adj.Date.IsZero() {
		adj.Date = time.Now()
	}

	var id string
	err = s.db.QueryRow(ctx, `
		INSERT INTO goal_adjustments (goal_id, user_id, amount, note, date, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id
	`, adj.GoalID, adj.UserID, adj.Amount, adj.Note, adj.Date).Scan(&id)
	if err != nil {
		return "", err
	}

	_ = s.UpdateGoalProgress(ctx, adj.UserID)
	return id, nil
}

// ListAdjustments retorna o histórico de ajustes de uma meta.
func (s *GoalsService) ListAdjustments(ctx context.Context, userID, goalID string) ([]GoalAdjustment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, goal_id, user_id, amount, COALESCE(note, ''), date, created_at
		FROM goal_adjustments
		WHERE goal_id = $1 AND user_id = $2
		ORDER BY date DESC, created_at DESC
	`, goalID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjustments []GoalAdjustment
	for rows.Next() {
		var a GoalAdjustment
		if err := rows.Scan(&a.ID, &a.GoalID, &a.UserID, &a.Amount, &a.Note, &a.Date, &a.CreatedAt); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, a)
	}
	return adjustments, nil
}

// UpdateGoalProgress recalcula determinística e idempotentemente o progresso de todas as metas ativas do usuário.
func (s *GoalsService) UpdateGoalProgress(ctx context.Context, userID string) error {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, goal_type, target_amount, initial_amount, start_date, target_date, category_id, account_id, installment_id, status
		FROM financial_goals
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type rawGoal struct {
		id            string
		name          string
		goalType      GoalType
		targetAmount  float64
		initialAmount float64
		startDate     time.Time
		targetDate    *time.Time
		categoryID    *string
		accountID     *string
		installmentID *string
		status        string
	}

	var goals []rawGoal
	for rows.Next() {
		var g rawGoal
		if err := rows.Scan(&g.id, &g.name, &g.goalType, &g.targetAmount, &g.initialAmount, &g.startDate, &g.targetDate, &g.categoryID, &g.accountID, &g.installmentID, &g.status); err != nil {
			return err
		}
		goals = append(goals, g)
	}

	for _, g := range goals {
		var autoAmount float64

		switch g.goalType {
		case GoalSpendingLimit:
			// 1. Limite de gastos: soma débitos da categoria no período
			start := g.startDate
			if g.targetDate == nil && start.Before(time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())) {
				// Se sem target_date e start_date é anterior ao mês atual, analisa o mês corrente
				start = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
			}
			s.db.QueryRow(ctx, `
				SELECT COALESCE(SUM(t.amount), 0)
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				WHERE a.user_id = $1 AND t.direction = 'debit'
				  AND ($2::uuid IS NULL OR t.category_id = $2)
				  AND t.date >= $3 AND ($4::date IS NULL OR t.date <= $4)
			`, userID, g.categoryID, start, g.targetDate).Scan(&autoAmount)

		case GoalIncomeTarget:
			// 2. Meta de renda: soma créditos no período
			start := g.startDate
			if g.targetDate == nil && start.Before(time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())) {
				start = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
			}
			s.db.QueryRow(ctx, `
				SELECT COALESCE(SUM(t.amount), 0)
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				WHERE a.user_id = $1 AND t.direction = 'credit'
				  AND t.date >= $2 AND ($3::date IS NULL OR t.date <= $3)
			`, userID, start, g.targetDate).Scan(&autoAmount)

		case GoalSavings:
			// 3. Economia / Reserva: evolução da conta vinculada ou saldo total desde initial_amount
			var currentBalance float64
			if g.accountID != nil {
				s.db.QueryRow(ctx, "SELECT COALESCE(balance, 0) FROM connected_accounts WHERE id = $1 AND user_id = $2", *g.accountID, userID).Scan(&currentBalance)
			} else {
				s.db.QueryRow(ctx, "SELECT COALESCE(SUM(balance), 0) FROM connected_accounts WHERE user_id = $1", userID).Scan(&currentBalance)
			}
			if g.initialAmount > 0 {
				autoAmount = math.Max(0, currentBalance-g.initialAmount)
			} else {
				autoAmount = currentBalance
			}

		case GoalDebtPayoff:
			// 4. Quitação de dívida: parcelas amortizadas do compromisso vinculado
			if g.installmentID != nil {
				var totalAmount float64
				var totalParts, currentPart int
				err := s.db.QueryRow(ctx, `
					SELECT i.total_amount, i.installments_total, i.installment_current
					FROM installments i
					JOIN connected_accounts a ON i.account_id = a.id
					WHERE i.id = $1 AND a.user_id = $2
				`, *g.installmentID, userID).Scan(&totalAmount, &totalParts, &currentPart)
				if err == nil && totalParts > 0 {
					partValue := totalAmount / float64(totalParts)
					autoAmount = float64(currentPart) * partValue
				}
			}
		}

		// 5. Soma os ajustes manuais da meta
		var adjustmentsTotal float64
		s.db.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0)
			FROM goal_adjustments
			WHERE goal_id = $1 AND user_id = $2
		`, g.id, userID).Scan(&adjustmentsTotal)

		totalCurrent := autoAmount + adjustmentsTotal
		if totalCurrent < 0 {
			totalCurrent = 0
		}

		// Atualiza o status
		status := g.status
		if status != string(GoalStatusPaused) {
			if g.goalType == GoalSpendingLimit {
				if totalCurrent > g.targetAmount {
					status = string(GoalStatusFailed)
				} else {
					status = string(GoalStatusActive)
				}
			} else {
				if totalCurrent >= g.targetAmount {
					status = string(GoalStatusCompleted)
				} else {
					status = string(GoalStatusActive)
				}
			}
		}

		s.db.Exec(ctx, `
			UPDATE financial_goals
			SET current_amount = $1, status = $2, updated_at = NOW()
			WHERE id = $3 AND user_id = $4
		`, totalCurrent, status, g.id, userID)
	}

	return nil
}

// GetPlanningTimeline unifica parcelas, assinaturas, renda prevista e prazos de metas em ordem cronológica.
func (s *GoalsService) GetPlanningTimeline(ctx context.Context, userID string) ([]PlanningTimelineItem, error) {
	var items []PlanningTimelineItem
	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())

	// 1. Parcelas ativas
	instRows, err := s.db.Query(ctx, `
		SELECT i.id, i.merchant_name, i.total_amount, i.installments_total, i.installment_current, i.start_date
		FROM installments i
		JOIN connected_accounts a ON i.account_id = a.id
		WHERE a.user_id = $1 AND i.installment_current < i.installments_total
	`, userID)
	if err == nil {
		defer instRows.Close()
		for instRows.Next() {
			var id, merchant string
			var totalAmount float64
			var totalParts, currentPart int
			var startDate time.Time
			if err := instRows.Scan(&id, &merchant, &totalAmount, &totalParts, &currentPart, &startDate); err == nil {
				partAmount := totalAmount / float64(totalParts)
				nextDue := startDate.AddDate(0, 1, 0)
				if !nextDue.Before(today) {
					items = append(items, PlanningTimelineItem{
						Type:      "installment",
						Title:     fmt.Sprintf("Parcela %s (%d/%d)", merchant, currentPart+1, totalParts),
						Amount:    partAmount,
						Direction: "debit",
						Date:      nextDue,
						Details: map[string]interface{}{
							"installment_id":      id,
							"merchant":            merchant,
							"installment_current": currentPart + 1,
							"installments_total":  totalParts,
							"remaining_amount":    partAmount * float64(totalParts-currentPart),
						},
					})
				}
			}
		}
	}

	// 2. Renda prevista recorrente (salário)
	var salaryAmount float64
	var salaryMerchant string
	var nextSalaryDate *time.Time
	err = s.db.QueryRow(ctx, `
		SELECT 
			COALESCE(NULLIF(t.merchant_name, ''), t.description),
			AVG(t.amount),
			MAX(t.date) + INTERVAL '1 month'
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'credit'
		  AND t.date >= DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '2 months'
		GROUP BY COALESCE(NULLIF(t.merchant_name, ''), t.description)
		HAVING BOOL_OR(LOWER(t.description) ~ '(sal[aá]rio|folha de pagamento|pr[oó][ -]?labore)')
		   OR (COUNT(DISTINCT DATE_TRUNC('month', t.date)) >= 2
		       AND COALESCE((MAX(t.amount) - MIN(t.amount)) / NULLIF(AVG(t.amount), 0), 1) <= 0.20)
		ORDER BY MAX(t.date) DESC
		LIMIT 1
	`, userID).Scan(&salaryMerchant, &salaryAmount, &nextSalaryDate)
	if err == nil && nextSalaryDate != nil && !nextSalaryDate.Before(today) {
		items = append(items, PlanningTimelineItem{
			Type:      "recurring_income",
			Title:     fmt.Sprintf("Renda prevista: %s", salaryMerchant),
			Amount:    salaryAmount,
			Direction: "credit",
			Date:      *nextSalaryDate,
			Details: map[string]interface{}{
				"source": salaryMerchant,
			},
		})
	}

	// 3. Prazos de Metas ativas
	goalRows, err := s.db.Query(ctx, `
		SELECT id, name, goal_type, target_amount, current_amount, target_date
		FROM financial_goals
		WHERE user_id = $1 AND status = 'active' AND target_date IS NOT NULL
	`, userID)
	if err == nil {
		defer goalRows.Close()
		for goalRows.Next() {
			var id, name string
			var goalType GoalType
			var targetAmount, currentAmount float64
			var targetDate time.Time
			if err := goalRows.Scan(&id, &name, &goalType, &targetAmount, &currentAmount, &targetDate); err == nil {
				if !targetDate.Before(today) {
					rem := targetAmount - currentAmount
					if rem < 0 {
						rem = 0
					}
					items = append(items, PlanningTimelineItem{
						Type:      "goal_deadline",
						Title:     fmt.Sprintf("Meta: %s", name),
						Amount:    rem,
						Direction: "info",
						Date:      targetDate,
						Details: map[string]interface{}{
							"goal_id":        id,
							"goal_type":      goalType,
							"target_amount":  targetAmount,
							"current_amount": currentAmount,
						},
					})
				}
			}
		}
	}

	// Ordena toda a timeline em ordem cronológica crescente
	sort.Slice(items, func(i, j int) bool {
		return items[i].Date.Before(items[j].Date)
	})

	return items, nil
}


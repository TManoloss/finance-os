package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GoalType string

const (
	GoalSavings       GoalType = "savings"
	GoalDebtPayoff    GoalType = "debt_payoff"
	GoalSpendingLimit GoalType = "spending_limit"
	GoalIncomeTarget  GoalType = "income_target"
)

type FinancialGoal struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	GoalType      GoalType  `json:"goal_type"`
	TargetAmount  float64   `json:"target_amount"`
	CurrentAmount float64   `json:"current_amount"`
	StartDate     time.Time `json:"start_date"`
	TargetDate    *time.Time `json:"target_date"`
	CategoryID    *string   `json:"category_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type GoalsService struct {
	db *pgxpool.Pool
}

func NewGoalsService(db *pgxpool.Pool) *GoalsService {
	return &GoalsService{db: db}
}

func (s *GoalsService) ListGoals(ctx context.Context, userID string) ([]FinancialGoal, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, name, goal_type, target_amount, current_amount, start_date, target_date, category_id, status, created_at
		FROM financial_goals WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []FinancialGoal
	for rows.Next() {
		var g FinancialGoal
		err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.GoalType, &g.TargetAmount, &g.CurrentAmount, &g.StartDate, &g.TargetDate, &g.CategoryID, &g.Status, &g.CreatedAt)
		if err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, nil
}

func (s *GoalsService) CreateGoal(ctx context.Context, g FinancialGoal) (string, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO financial_goals (user_id, name, goal_type, target_amount, start_date, target_date, category_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, g.UserID, g.Name, g.GoalType, g.TargetAmount, g.StartDate, g.TargetDate, g.CategoryID).Scan(&id)
	return id, err
}

// UpdateGoalProgress recalcula o progresso de todas as metas ativas do usuário.
func (s *GoalsService) UpdateGoalProgress(ctx context.Context, userID string) error {
	goals, err := s.ListGoals(ctx, userID)
	if err != nil {
		return err
	}

	for _, g := range goals {
		if g.Status != "active" {
			continue
		}

		var currentAmount float64
		switch g.GoalType {
		case GoalSpendingLimit:
			// Soma gastos na categoria no mês atual
			if g.CategoryID != nil {
				s.db.QueryRow(ctx, `
					SELECT COALESCE(SUM(amount), 0)
					FROM transactions t
					JOIN connected_accounts a ON t.account_id = a.id
					WHERE a.user_id = $1 AND t.category_id = $2 
					  AND t.direction = 'debit'
					  AND EXTRACT(MONTH FROM t.date) = EXTRACT(MONTH FROM NOW())
					  AND EXTRACT(YEAR FROM t.date) = EXTRACT(YEAR FROM NOW())
				`, userID, *g.CategoryID).Scan(&currentAmount)
			}
		case GoalSavings:
			// Soma de saldo economizado (Simplificado: saldo atual - saldo na data de início)
			// Requereria histórico de saldos, vamos usar saldo atual por enquanto
			s.db.QueryRow(ctx, "SELECT COALESCE(SUM(balance), 0) FROM connected_accounts WHERE user_id = $1", userID).Scan(&currentAmount)
		}

		// Atualiza o progresso
		status := "active"
		if g.GoalType == GoalSpendingLimit && currentAmount > g.TargetAmount {
			status = "failed"
		} else if g.GoalType != GoalSpendingLimit && currentAmount >= g.TargetAmount {
			status = "completed"
		}

		s.db.Exec(ctx, "UPDATE financial_goals SET current_amount = $1, status = $2 WHERE id = $3", currentAmount, status, g.ID)
	}

	return nil
}

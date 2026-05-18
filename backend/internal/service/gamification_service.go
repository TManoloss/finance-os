package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Mission struct {
	ID           string     `json:"id"`
	TemplateID   string     `json:"template_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	TargetValue  float64    `json:"target_value"`
	CurrentValue float64    `json:"current_value"`
	StartedAt    time.Time  `json:"started_at"`
	EndsAt       *time.Time `json:"ends_at"`
	Status       string     `json:"status"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type Achievement struct {
	ID            string      `json:"id"`
	AchievementID string      `json:"achievement_id"`
	AwardedAt     time.Time   `json:"awarded_at"`
	ContextData   interface{} `json:"context_data"`
}

type SalaryPlan struct {
	ID                string      `json:"id"`
	SalaryDetected    float64     `json:"salary_detected"`
	FixedCommitments  float64     `json:"fixed_commitments"`
	SafeDailyLimit    float64     `json:"safe_daily_limit"`
	PlanData          interface{} `json:"plan_data"`
	ValidUntil        time.Time   `json:"valid_until"`
	GeneratedAt       time.Time   `json:"generated_at"`
}

type GamificationService struct {
	db *pgxpool.Pool
}

func NewGamificationService(db *pgxpool.Pool) *GamificationService {
	return &GamificationService{db: db}
}

// GetActiveMissions retorna as missões ativas de um usuário.
func (s *GamificationService) GetActiveMissions(ctx context.Context, userID string) ([]Mission, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, template_id, title, description, target_value, current_value, started_at, ends_at, status, completed_at
		FROM missions
		WHERE user_id = $1 AND status = 'active'
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var missions []Mission
	for rows.Next() {
		var m Mission
		err := rows.Scan(
			&m.ID, &m.TemplateID, &m.Title, &m.Description, &m.TargetValue, &m.CurrentValue,
			&m.StartedAt, &m.EndsAt, &m.Status, &m.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		missions = append(missions, m)
	}
	return missions, nil
}

// GetAwardedAchievements retorna as conquistas recebidas por um usuário.
func (s *GamificationService) GetAwardedAchievements(ctx context.Context, userID string) ([]Achievement, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, achievement_id, awarded_at, context_data
		FROM achievements_awarded
		WHERE user_id = $1
		ORDER BY awarded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []Achievement
	for rows.Next() {
		var a Achievement
		err := rows.Scan(&a.ID, &a.AchievementID, &a.AwardedAt, &a.ContextData)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}
	return achievements, nil
}

// GetSalaryPlan retorna o plano de salário mais recente e válido do usuário.
func (s *GamificationService) GetSalaryPlan(ctx context.Context, userID string) (*SalaryPlan, error) {
	var p SalaryPlan
	err := s.db.QueryRow(ctx, `
		SELECT id, salary_detected, fixed_commitments, safe_daily_limit, plan_data, valid_until, generated_at
		FROM salary_plans
		WHERE user_id = $1 AND valid_until >= CURRENT_DATE
		ORDER BY generated_at DESC
		LIMIT 1
	`, userID).Scan(&p.ID, &p.SalaryDetected, &p.FixedCommitments, &p.SafeDailyLimit, &p.PlanData, &p.ValidUntil, &p.GeneratedAt)
	
	if err != nil {
		return nil, err
	}
	return &p, nil
}

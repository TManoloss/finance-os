package service

import (
	"context"
	"sort"
	"testing"
	"time"
)

func TestGoalTypes(t *testing.T) {
	types := []GoalType{GoalSavings, GoalDebtPayoff, GoalSpendingLimit, GoalIncomeTarget}
	expected := []string{"savings", "debt_payoff", "spending_limit", "income_target"}

	for i, gt := range types {
		if string(gt) != expected[i] {
			t.Errorf("GoalType[%d] = %s, want %s", i, gt, expected[i])
		}
	}
}

func TestGoalStatuses(t *testing.T) {
	statuses := []GoalStatus{GoalStatusActive, GoalStatusPaused, GoalStatusCompleted, GoalStatusFailed}
	expected := []string{"active", "paused", "completed", "failed"}

	for i, gs := range statuses {
		if string(gs) != expected[i] {
			t.Errorf("GoalStatus[%d] = %s, want %s", i, gs, expected[i])
		}
	}
}

func TestCreateGoalRejectsInvalidFieldsBeforeDatabaseAccess(t *testing.T) {
	start := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	targetBeforeStart := start.AddDate(0, 0, -1)

	for name, test := range map[string]struct {
		goal FinancialGoal
		want string
	}{
		"goal type": {
			goal: FinancialGoal{
				Name:         "Reserva",
				GoalType:     GoalType("unknown"),
				TargetAmount: 100,
				StartDate:    start,
				Status:       string(GoalStatusActive),
			},
			want: "tipo de meta inválido",
		},
		"status": {
			goal: FinancialGoal{
				Name:         "Reserva",
				GoalType:     GoalSavings,
				TargetAmount: 100,
				StartDate:    start,
				Status:       "unknown",
			},
			want: "status de meta inválido",
		},
		"target date": {
			goal: FinancialGoal{
				Name:         "Reserva",
				GoalType:     GoalSavings,
				TargetAmount: 100,
				StartDate:    start,
				TargetDate:   &targetBeforeStart,
				Status:       string(GoalStatusActive),
			},
			want: "data alvo não pode ser anterior ao início",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewGoalsService(nil).CreateGoal(context.Background(), test.goal)
			if err == nil || err.Error() != test.want {
				t.Fatalf("CreateGoal() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestPlanningTimelineSorting(t *testing.T) {
	now := time.Now()
	items := []PlanningTimelineItem{
		{Title: "Item 3", Date: now.AddDate(0, 0, 15)},
		{Title: "Item 1", Date: now.AddDate(0, 0, 2)},
		{Title: "Item 2", Date: now.AddDate(0, 0, 7)},
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Date.Before(items[j].Date)
	})

	if items[0].Title != "Item 1" || items[1].Title != "Item 2" || items[2].Title != "Item 3" {
		t.Errorf("Timeline items not sorted chronologically: got %+v", items)
	}
}

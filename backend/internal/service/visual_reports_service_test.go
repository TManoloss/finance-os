package service

import "testing"

func TestReplayOutcome(t *testing.T) {
	tests := []struct {
		name       string
		netSavings float64
		totalSpent float64
		previous   float64
		category   string
		wantResult string
		wantAdvice string
	}{
		{
			name:       "improvement",
			netSavings: 100,
			totalSpent: 900,
			previous:   1000,
			wantResult: "improvement",
			wantAdvice: "Mantenha o ritmo atual e acompanhe se a redução continua no próximo mês.",
		},
		{
			name:       "deficit setback with category",
			netSavings: -50,
			totalSpent: 1000,
			previous:   900,
			category:   "Alimentação",
			wantResult: "setback",
			wantAdvice: "Revise primeiro os gastos em Alimentação e preserve saldo para os compromissos do próximo mês.",
		},
		{
			name:       "insufficient history",
			wantResult: "insufficient_history",
			wantAdvice: "Continue acompanhando o ritmo semanal até existir histórico comparável suficiente.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotAdvice := replayOutcome(tt.netSavings, tt.totalSpent, tt.previous, tt.category)
			if gotResult != tt.wantResult || gotAdvice != tt.wantAdvice {
				t.Fatalf("replayOutcome() = (%q, %q), want (%q, %q)", gotResult, gotAdvice, tt.wantResult, tt.wantAdvice)
			}
		})
	}
}

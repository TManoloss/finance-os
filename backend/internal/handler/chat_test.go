package handler

import (
	"testing"

	"github.com/finance-os/backend/internal/service"
)

func TestDeterministicChatFallbackIncludesCalculatedContext(t *testing.T) {
	got := deterministicChatFallback(
		&service.ConsolidatedIntelligence{
			OverallHealthScore: 72.4,
			HealthStatus:       "good",
			Quality:            "high",
		},
		&service.MonthlyReplayResult{
			Insights:          []string{"Seu gasto caiu."},
			NextMonthGuidance: "Acompanhe o próximo mês.",
		},
	)

	want := "A explicação por IA está indisponível agora, mas os cálculos do FinanceOS continuam funcionando.\n\n" +
		"Sua saúde financeira está em **72/100** (boa), com qualidade de dados alta.\n\n" +
		"Seu gasto caiu.\n\nAcompanhe o próximo mês."
	if got != want {
		t.Fatalf("deterministicChatFallback() = %q, want %q", got, want)
	}
}

func TestDeterministicChatFallbackDoesNotInventData(t *testing.T) {
	const want = "A explicação por IA está indisponível agora, mas os cálculos do FinanceOS continuam funcionando."

	if got := deterministicChatFallback(nil, nil); got != want {
		t.Fatalf("deterministicChatFallback(nil, nil) = %q, want %q", got, want)
	}
}

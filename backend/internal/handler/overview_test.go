package handler

import (
	"testing"
	"time"

	"github.com/finance-os/backend/internal/service"
)

func TestOverviewSelectsUnreadAlertAndNearestCommitment(t *testing.T) {
	now := time.Now()
	readAt := now
	alert := selectMainAlert([]service.FeedEvent{
		{ID: "read", Severity: "alert", ReadAt: &readAt},
		{ID: "warning", Severity: "warning"},
		{ID: "info", Severity: "info"},
	})
	if alert == nil || alert.ID != "warning" {
		t.Fatalf("alerta principal = %#v", alert)
	}

	commitment := selectNextCommitment(
		[]service.ActiveInstallment{{MerchantName: "Parcela", Amount: 90, NextDueDate: now.Add(48 * time.Hour)}},
		[]service.Subscription{{MerchantName: "Assinatura", Amount: 30, NextEstimate: now.Add(24 * time.Hour), Status: "active"}},
		now,
	)
	if commitment == nil || commitment.Title != "Assinatura" {
		t.Fatalf("próximo compromisso = %#v", commitment)
	}
}

func TestOverviewDoesNotReturnExpiredCommitment(t *testing.T) {
	today := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	commitment := selectNextCommitment(nil, []service.Subscription{{
		MerchantName: "Pagamento de fatura",
		NextEstimate: time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC),
		Status:       "active",
	}}, today)
	if commitment == nil || commitment.DueDate.Before(today) {
		t.Fatalf("próximo compromisso vencido = %#v", commitment)
	}
}

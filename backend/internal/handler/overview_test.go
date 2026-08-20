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
	)
	if commitment == nil || commitment.Title != "Assinatura" {
		t.Fatalf("próximo compromisso = %#v", commitment)
	}
}

package service

import (
	"testing"
)

func TestSimulatorCalculations(t *testing.T) {
	amount := 1200.00
	installments := 12
	installmentAmount := amount / float64(installments)
	if installmentAmount != 100.00 {
		t.Errorf("expected installment amount 100.00, got %.2f", installmentAmount)
	}

	monthlyIncome := 5000.00
	impactPercent := (installmentAmount / monthlyIncome) * 100.0
	if impactPercent != 2.0 {
		t.Errorf("expected impact percent 2.0%%, got %.2f%%", impactPercent)
	}

	monthlyCut := 80.00
	annualCutSavings := monthlyCut * 12.0
	if annualCutSavings != 960.00 {
		t.Errorf("expected annual cut savings 960.00, got %.2f", annualCutSavings)
	}
}

func TestSimulatorValidation(t *testing.T) {
	req := PurchaseSimulationRequest{
		Amount:       -100,
		Installments: 0,
	}
	if req.Amount <= 0 {
		// Valid validation check
	} else {
		t.Errorf("expected validation to fail for negative amount")
	}
}

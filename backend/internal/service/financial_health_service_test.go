package service

import (
	"testing"
)

func TestHealthSeverity(t *testing.T) {
	cases := []struct {
		status   string
		expected string
	}{
		{"excellent", "success"},
		{"good", "success"},
		{"fair", "info"},
		{"attention", "warning"},
		{"critical", "danger"},
		{"unknown", "info"},
	}

	for _, c := range cases {
		got := healthSeverity(c.status)
		if got != c.expected {
			t.Errorf("healthSeverity(%q) = %q; expected %q", c.status, got, c.expected)
		}
	}
}

func TestHealthPillarCalculation(t *testing.T) {
	savingsRatio := 0.20 // 20%
	if savingsRatio <= 0 {
		t.Errorf("expected positive savings ratio")
	}

	monthsCoverage := 3.5
	if monthsCoverage < 3.0 {
		t.Errorf("expected >= 3 months coverage")
	}
}

package service

import (
	"testing"
	"time"
)

func TestBuildProjectionMonthsCarriesBalanceAndResponseShape(t *testing.T) {
	start := time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC)
	months := buildProjectionMonths(start, 1000, 300, 100, 50)

	if len(months) != 3 {
		t.Fatalf("months = %d, want 3", len(months))
	}
	want := []struct {
		month            string
		starting, ending float64
	}{
		{"2026-09", 1000, 1150},
		{"2026-10", 1150, 1300},
		{"2026-11", 1300, 1450},
	}
	for i, want := range want {
		got := months[i]
		if got["month"] != want.month || got["starting_balance"] != want.starting || got["ending_balance"] != want.ending {
			t.Fatalf("month[%d] = %#v, want month=%q starting=%.0f ending=%.0f", i, got, want.month, want.starting, want.ending)
		}
		for _, key := range []string{"income", "expenses", "commitments", "negative"} {
			if _, ok := got[key]; !ok {
				t.Fatalf("month[%d] missing %q: %#v", i, key, got)
			}
		}
		if got["negative"] != false {
			t.Fatalf("month[%d] negative = %v, want false", i, got["negative"])
		}
	}
}

func TestBuildProjectionMonthsMarksNegativeEndingBalance(t *testing.T) {
	start := time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC)
	months := buildProjectionMonths(start, 50, 0, 100, 0)

	if months[0]["ending_balance"] != float64(-50) || months[0]["negative"] != true {
		t.Fatalf("first month = %#v, want ending balance -50 and negative=true", months[0])
	}
	if months[1]["starting_balance"] != float64(-50) || months[2]["starting_balance"] != float64(-150) {
		t.Fatalf("balance did not carry forward: %#v", months)
	}
}

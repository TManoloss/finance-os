package handler

import (
	"testing"
	"time"
)

func TestNextDailySync(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}

	before := time.Date(2026, 8, 18, 0, 29, 0, 0, loc)
	if got := nextDailySync(before); got.Day() != 18 || got.Hour() != 0 || got.Minute() != 30 {
		t.Fatalf("unexpected same-day schedule: %s", got)
	}

	atSchedule := time.Date(2026, 8, 18, 0, 30, 0, 0, loc)
	if got := nextDailySync(atSchedule); got.Day() != 19 || got.Hour() != 0 || got.Minute() != 30 {
		t.Fatalf("unexpected next-day schedule: %s", got)
	}
}

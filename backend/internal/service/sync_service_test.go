package service

import "testing"

func TestLastFour(t *testing.T) {
	for input, want := range map[string]string{"": "", "123": "123", "12345678": "5678"} {
		if got := lastFour(input); got != want {
			t.Fatalf("lastFour(%q) = %q, quer %q", input, got, want)
		}
	}
}

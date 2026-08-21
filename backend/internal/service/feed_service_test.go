package service

import "testing"

func TestUnusualSpendingThreshold(t *testing.T) {
	tests := []struct {
		name   string
		median float64
		want   float64
	}{
		{name: "mediana zero", median: 0, want: 1000},
		{name: "mediana abaixo do piso", median: 100, want: 1000},
		{name: "tres vezes a mediana", median: 400, want: 1200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unusualSpendingThreshold(tt.median); got != tt.want {
				t.Errorf("unusualSpendingThreshold(%v) = %v, want %v", tt.median, got, tt.want)
			}
		})
	}
}

func TestIsSalaryDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        bool
	}{
		{name: "salário com acento", description: "Salário mensal", want: true},
		{name: "salario sem acento", description: "Salario mensal", want: true},
		{name: "folha de pagamento", description: "Folha de pagamento", want: true},
		{name: "pro-labore", description: "Pró-labore", want: true},
		{name: "descrição comum", description: "Compra no supermercado", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSalaryDescription(tt.description); got != tt.want {
				t.Errorf("isSalaryDescription(%q) = %v, want %v", tt.description, got, tt.want)
			}
		})
	}
}

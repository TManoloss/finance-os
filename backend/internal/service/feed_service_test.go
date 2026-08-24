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

func TestFormatSubscriptionChange(t *testing.T) {
	title, desc, severity := formatSubscriptionChange("Netflix", 39.90, 55.90)
	if title != "Aumento na assinatura 💳" {
		t.Errorf("expected title 'Aumento na assinatura 💳', got %q", title)
	}
	if severity != "warning" {
		t.Errorf("expected severity 'warning', got %q", severity)
	}
	if desc != "A cobrança de Netflix passou de R$ 39,90 para R$ 55,90 (+R$ 16,00)." {
		t.Errorf("unexpected description: %q", desc)
	}

	titleRed, descRed, severityRed := formatSubscriptionChange("Spotify", 34.90, 21.90)
	if titleRed != "Redução na assinatura 🎉" {
		t.Errorf("expected title 'Redução na assinatura 🎉', got %q", titleRed)
	}
	if severityRed != "info" {
		t.Errorf("expected severity 'info', got %q", severityRed)
	}
	if descRed != "A cobrança de Spotify reduziu de R$ 34,90 para R$ 21,90 (-R$ 13,00)." {
		t.Errorf("unexpected description: %q", descRed)
	}
}

func TestFormatMonthlyClose(t *testing.T) {
	title, desc := formatMonthlyClose("Julho", 1500.50, 3000.00)
	if title != "Fechamento de Julho 📊" {
		t.Errorf("expected title 'Fechamento de Julho 📊', got %q", title)
	}
	if desc != "No mês anterior, você teve R$ 1.500,50 em gastos e R$ 3.000,00 em entradas (resultado líquido: +R$ 1.499,50)." {
		t.Errorf("unexpected description: %q", desc)
	}

	_, descDeficit := formatMonthlyClose("Agosto", 4000.00, 3000.00)
	if descDeficit != "No mês anterior, você teve R$ 4.000,00 em gastos e R$ 3.000,00 em entradas (resultado líquido: R$ -1.000,00)." {
		t.Errorf("unexpected description: %q", descDeficit)
	}
}

func TestFormatCategorySpike(t *testing.T) {
	title, desc := formatCategorySpike("Alimentação", 450.00, 200.00, 125.0)
	if title != "Aumento em Alimentação 📈" {
		t.Errorf("expected title 'Aumento em Alimentação 📈', got %q", title)
	}
	if desc != "Seus gastos com Alimentação este mês (R$ 450,00) já superam em 125% o total do mês anterior (R$ 200,00)." {
		t.Errorf("unexpected description: %q", desc)
	}
}

func TestPortugueseMonthName(t *testing.T) {
	if got := portugueseMonthName(1); got != "Janeiro" {
		t.Errorf("got %q, want 'Janeiro'", got)
	}
	if got := portugueseMonthName(12); got != "Dezembro" {
		t.Errorf("got %q, want 'Dezembro'", got)
	}
}

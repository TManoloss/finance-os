package handler

import (
	"fmt"
	"net/http"

	"github.com/finance-os/backend/internal/response"
	"github.com/labstack/echo/v4"
)

type SimulatorHandler struct{}

func NewSimulatorHandler() *SimulatorHandler {
	return &SimulatorHandler{}
}

// PurchaseRequest simula o impacto de uma compra parcelada.
type PurchaseRequest struct {
	Amount       float64 `json:"amount"`
	Installments int     `json:"installments"`
	Description  string  `json:"description"`
}

func (h *SimulatorHandler) SimulatePurchase(c echo.Context) error {
	var req PurchaseRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "body inválido")
	}

	monthly := req.Amount / float64(req.Installments)
	annual := monthly * 12
	if float64(req.Installments) < 12 {
		annual = req.Amount
	}

	// Simular projeção de saldo (mockado por enquanto, mas com a estrutura certa)
	projections := []map[string]interface{}{}
	months := []string{"JAN", "FEV", "MAR", "ABR", "MAI", "JUN"}
	for i, m := range months {
		balance := 1200.0 - (monthly * float64(i+1))
		projections = append(projections, map[string]interface{}{
			"month":   m,
			"balance": balance,
		})
	}

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"monthly_impact": monthly,
		"annual_impact":  annual,
		"projections":    projections,
		"risk_alerts": []map[string]string{
			{"message": "Esta compra compromete 15% da sua renda mensal disponível."},
		},
	})
}

// CutSubscriptionRequest simula a economia ao cortar uma assinatura.
type CutSubscriptionRequest struct {
	Amount       float64 `json:"monthly_amount"`
	MerchantName string  `json:"merchant_name"`
}

func (h *SimulatorHandler) SimulateCut(c echo.Context) error {
	var req CutSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "body inválido")
	}

	annual := req.Amount * 12
	fiveYears := annual * 5
	invested := fiveYears * 1.45

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"monthly_impact": req.Amount,
		"annual_impact":  annual,
		"opportunity_cost": map[string]interface{}{
			"narrative":      fmt.Sprintf("Ao cortar %s, você economiza R$ %.2f em 5 anos.", "este serviço", fiveYears),
			"invested_value": invested,
		},
	})
}

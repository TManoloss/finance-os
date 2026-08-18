package handler

import (
	"net/http"

	"github.com/finance-os/backend/internal/response"
	"github.com/labstack/echo/v4"
)

type SimulatorHandler struct{}

func NewSimulatorHandler() *SimulatorHandler {
	return &SimulatorHandler{}
}

func (h *SimulatorHandler) SimulatePurchase(c echo.Context) error {
	return response.Error(c, http.StatusServiceUnavailable, "simulador indisponível até usar dados financeiros reais")
}

func (h *SimulatorHandler) SimulateCut(c echo.Context) error {
	return response.Error(c, http.StatusServiceUnavailable, "simulador indisponível até usar dados financeiros reais")
}

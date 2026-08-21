package handler

import (
	"net/http"

	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type SimulatorHandler struct {
	service *service.SimulatorService
}

func NewSimulatorHandler(s *service.SimulatorService) *SimulatorHandler {
	return &SimulatorHandler{service: s}
}

// SimulatePurchase projeta o impacto financeiro real de uma compra.
func (h *SimulatorHandler) SimulatePurchase(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req service.PurchaseSimulationRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "corpo da requisição inválido")
	}

	result, err := h.service.SimulatePurchase(c.Request().Context(), userID, req)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	return response.Success(c, http.StatusOK, result)
}

// SimulateCut projeta o impacto financeiro do corte de uma despesa sem rentabilidade fictícia.
func (h *SimulatorHandler) SimulateCut(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req service.CutSimulationRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "corpo da requisição inválido")
	}

	result, err := h.service.SimulateCut(c.Request().Context(), userID, req)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	return response.Success(c, http.StatusOK, result)
}

// SaveSimulation persiste uma simulação com nome e parâmetros.
func (h *SimulatorHandler) SaveSimulation(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		SimulationType string                 `json:"simulation_type"`
		Name           string                 `json:"name"`
		InputParams    map[string]interface{} `json:"input_params"`
		ResultJSON     map[string]interface{} `json:"result_json"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "corpo da requisição inválido")
	}

	id, err := h.service.SaveSimulation(c.Request().Context(), userID, req.SimulationType, req.Name, req.InputParams, req.ResultJSON)
	if err != nil {
		if err.Error() == "tipo de simulação inválido" || err.Error() == "parâmetros e resultado da simulação são obrigatórios" {
			return response.Error(c, http.StatusBadRequest, err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "erro ao salvar simulação")
	}

	return response.Success(c, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"message": "simulação salva com sucesso",
	})
}

// ListSaved lista as simulações salvas do usuário.
func (h *SimulatorHandler) ListSaved(c echo.Context) error {
	userID := c.Get("user_id").(string)

	sims, err := h.service.ListSavedSimulations(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao listar simulações salvas")
	}

	return response.Success(c, http.StatusOK, sims)
}

// DeleteSaved remove uma simulação salva do usuário.
func (h *SimulatorHandler) DeleteSaved(c echo.Context) error {
	userID := c.Get("user_id").(string)
	simID := c.Param("id")

	deleted, err := h.service.DeleteSavedSimulation(c.Request().Context(), userID, simID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao excluir simulação")
	}
	if !deleted {
		return response.Error(c, http.StatusNotFound, "simulação não encontrada")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"message": "simulação excluída com sucesso",
	})
}

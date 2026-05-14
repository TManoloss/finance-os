package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type GoalsHandler struct {
	service *service.GoalsService
	cfg     *config.Config
}

func NewGoalsHandler(s *service.GoalsService, cfg *config.Config) *GoalsHandler {
	return &GoalsHandler{service: s, cfg: cfg}
}

func (h *GoalsHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)
	goals, err := h.service.ListGoals(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao listar metas")
	}
	return response.Success(c, http.StatusOK, goals)
}

func (h *GoalsHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(string)
	var g service.FinancialGoal
	if err := c.Bind(&g); err != nil {
		return response.Error(c, http.StatusBadRequest, "body inválido")
	}
	g.UserID = userID
	g.StartDate = time.Now()
	
	id, err := h.service.CreateGoal(c.Request().Context(), g)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao criar meta")
	}
	g.ID = id
	return response.Success(c, http.StatusCreated, g)
}

func (h *GoalsHandler) Suggest(c echo.Context) error {
	userID := c.Get("user_id").(string)
	url := fmt.Sprintf("%s/agents/goals/suggest/%s", h.cfg.AgentsServiceURL, userID)
	
	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var suggestions interface{}
	json.NewDecoder(resp.Body).Decode(&suggestions)
	return response.Success(c, http.StatusOK, suggestions)
}

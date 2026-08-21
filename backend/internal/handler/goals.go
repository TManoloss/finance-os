package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// List lista todas as metas do usuário.
func (h *GoalsHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)
	goals, err := h.service.ListGoals(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao listar metas")
	}
	return response.Success(c, http.StatusOK, goals)
}

// Get retorna uma meta específica do usuário com histórico de ajustes.
func (h *GoalsHandler) Get(c echo.Context) error {
	userID := c.Get("user_id").(string)
	goalID := c.Param("id")

	goal, err := h.service.GetGoal(c.Request().Context(), userID, goalID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar meta")
	}
	if goal == nil {
		return response.Error(c, http.StatusNotFound, "meta não encontrada")
	}
	return response.Success(c, http.StatusOK, goal)
}

// Create cria uma nova meta.
func (h *GoalsHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(string)
	var req struct {
		Name          string  `json:"name"`
		GoalType      string  `json:"goal_type"`
		TargetAmount  float64 `json:"target_amount"`
		InitialAmount float64 `json:"initial_amount"`
		StartDate     string  `json:"start_date"`
		TargetDate    *string `json:"target_date"`
		CategoryID    *string `json:"category_id"`
		AccountID     *string `json:"account_id"`
		InstallmentID *string `json:"installment_id"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "corpo da requisição inválido")
	}

	if req.Name == "" || req.TargetAmount <= 0 {
		return response.Error(c, http.StatusBadRequest, "nome e valor alvo (> 0) são obrigatórios")
	}

	startDate := time.Now()
	if req.StartDate != "" {
		parsed, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return response.Error(c, http.StatusBadRequest, "data inicial inválida")
		}
		startDate = parsed
	}

	var targetDate *time.Time
	if req.TargetDate != nil && *req.TargetDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.TargetDate)
		if err != nil {
			return response.Error(c, http.StatusBadRequest, "data alvo inválida")
		}
		targetDate = &parsed
		if parsed.Before(startDate) {
			return response.Error(c, http.StatusBadRequest, "data alvo não pode ser anterior ao início")
		}
	}

	g := service.FinancialGoal{
		UserID:        userID,
		Name:          req.Name,
		GoalType:      service.GoalType(req.GoalType),
		TargetAmount:  req.TargetAmount,
		InitialAmount: req.InitialAmount,
		StartDate:     startDate,
		TargetDate:    targetDate,
		CategoryID:    req.CategoryID,
		AccountID:     req.AccountID,
		InstallmentID: req.InstallmentID,
		Status:        string(service.GoalStatusActive),
	}

	id, err := h.service.CreateGoal(c.Request().Context(), g)
	if err != nil {
		message := "erro ao criar meta"
		if strings.Contains(err.Error(), "inválid") || strings.Contains(err.Error(), "não encontrada") || strings.Contains(err.Error(), "obrigatório") || strings.Contains(err.Error(), "maior") {
			message = err.Error()
			return response.Error(c, http.StatusBadRequest, message)
		}
		return response.Error(c, http.StatusInternalServerError, message)
	}
	g.ID = id
	return response.Success(c, http.StatusCreated, g)
}

// Update atualiza propriedades de uma meta (nome, alvo, status, data).
func (h *GoalsHandler) Update(c echo.Context) error {
	userID := c.Get("user_id").(string)
	goalID := c.Param("id")

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	updated, err := h.service.UpdateGoal(c.Request().Context(), userID, goalID, updates)
	if err != nil {
		if err.Error() == "status de meta inválido" || err.Error() == "data alvo inválida" {
			return response.Error(c, http.StatusBadRequest, err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "erro ao atualizar meta")
	}
	if !updated {
		return response.Error(c, http.StatusNotFound, "meta não encontrada")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"message": "meta atualizada com sucesso",
	})
}

// Delete remove uma meta do usuário.
func (h *GoalsHandler) Delete(c echo.Context) error {
	userID := c.Get("user_id").(string)
	goalID := c.Param("id")

	deleted, err := h.service.DeleteGoal(c.Request().Context(), userID, goalID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao excluir meta")
	}
	if !deleted {
		return response.Error(c, http.StatusNotFound, "meta não encontrada")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"message": "meta excluída com sucesso",
	})
}

// AddAdjustment adiciona um aporte/ajuste manual a uma meta.
func (h *GoalsHandler) AddAdjustment(c echo.Context) error {
	userID := c.Get("user_id").(string)
	goalID := c.Param("id")

	var req struct {
		Amount float64 `json:"amount"`
		Note   string  `json:"note"`
		Date   string  `json:"date"`
	}

	if err := c.Bind(&req); err != nil || req.Amount == 0 {
		return response.Error(c, http.StatusBadRequest, "valor de ajuste (diferente de zero) é obrigatório")
	}

	adjDate := time.Now()
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return response.Error(c, http.StatusBadRequest, "data do ajuste inválida")
		}
		adjDate = parsed
	}

	adj := service.GoalAdjustment{
		GoalID: goalID,
		UserID: userID,
		Amount: req.Amount,
		Note:   req.Note,
		Date:   adjDate,
	}

	id, err := h.service.AddAdjustment(c.Request().Context(), adj)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao registrar ajuste")
	}
	adj.ID = id
	return response.Success(c, http.StatusCreated, adj)
}

// ListAdjustments lista os ajustes de uma meta.
func (h *GoalsHandler) ListAdjustments(c echo.Context) error {
	userID := c.Get("user_id").(string)
	goalID := c.Param("id")

	adjustments, err := h.service.ListAdjustments(c.Request().Context(), userID, goalID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao listar ajustes")
	}
	return response.Success(c, http.StatusOK, adjustments)
}

// GetTimeline retorna a linha temporal unificada de planejamento (parcelas, assinaturas, renda e metas).
func (h *GoalsHandler) GetTimeline(c echo.Context) error {
	userID := c.Get("user_id").(string)
	timeline, err := h.service.GetPlanningTimeline(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao obter linha temporal de planejamento")
	}
	return response.Success(c, http.StatusOK, timeline)
}

// Recalculate força o recálculo de todas as metas ativas do usuário.
func (h *GoalsHandler) Recalculate(c echo.Context) error {
	userID := c.Get("user_id").(string)
	if err := h.service.UpdateGoalProgress(c.Request().Context(), userID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao recalcular metas")
	}
	return response.Success(c, http.StatusOK, map[string]string{
		"message": "metas recalculadas com sucesso",
	})
}

// Suggest solicita sugestões de metas ao serviço de agentes.
func (h *GoalsHandler) Suggest(c echo.Context) error {
	userID := c.Get("user_id").(string)
	url := fmt.Sprintf("%s/agents/goals/suggest/%s", h.cfg.AgentsServiceURL, userID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return response.Error(c, http.StatusServiceUnavailable, "sugestões indisponíveis no momento")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return response.Error(c, http.StatusServiceUnavailable, "sugestões indisponíveis no momento")
	}

	var suggestions interface{}
	if err := json.NewDecoder(resp.Body).Decode(&suggestions); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar sugestões")
	}
	return response.Success(c, http.StatusOK, suggestions)
}

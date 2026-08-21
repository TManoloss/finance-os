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
		Name          string   `json:"name"`
		GoalType      string   `json:"goal_type"`
		TargetAmount  float64  `json:"target_amount"`
		InitialAmount float64  `json:"initial_amount"`
		StartDate     string   `json:"start_date"`
		TargetDate    *string  `json:"target_date"`
		CategoryID    *string  `json:"category_id"`
		AccountID     *string  `json:"account_id"`
		InstallmentID *string  `json:"installment_id"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "corpo da requisição inválido")
	}

	if req.Name == "" || req.TargetAmount <= 0 {
		return response.Error(c, http.StatusBadRequest, "nome e valor alvo (> 0) são obrigatórios")
	}

	startDate := time.Now()
	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = parsed
		}
	}

	var targetDate *time.Time
	if req.TargetDate != nil && *req.TargetDate != "" {
		if parsed, err := time.Parse("2006-01-02", *req.TargetDate); err == nil {
			targetDate = &parsed
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
		return response.Error(c, http.StatusInternalServerError, "erro ao criar meta: "+err.Error())
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
		if parsed, err := time.Parse("2006-01-02", req.Date); err == nil {
			adjDate = parsed
		}
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
		return response.Error(c, http.StatusInternalServerError, "erro ao registrar ajuste: "+err.Error())
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
		// Fallback amigável se o serviço de agentes estiver indisponível
		return response.Success(c, http.StatusOK, []map[string]interface{}{
			{
				"name":          "Reserva de Emergência",
				"goal_type":     "savings",
				"target_amount": 3000.00,
				"description":   "Construa uma reserva equivalente a 3 meses de gastos básicos.",
			},
			{
				"name":          "Controle de Alimentação",
				"goal_type":     "spending_limit",
				"target_amount": 600.00,
				"description":   "Defina um teto mensal para compras de alimentação e delivery.",
			},
		})
	}
	defer resp.Body.Close()

	var suggestions interface{}
	if err := json.NewDecoder(resp.Body).Decode(&suggestions); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar sugestões")
	}
	return response.Success(c, http.StatusOK, suggestions)
}


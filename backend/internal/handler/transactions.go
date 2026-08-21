package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type TransactionsHandler struct {
	repo       repository.TransactionRepository
	classifier *service.ClassifierService
	goals      *service.GoalsService
}

func (h *TransactionsHandler) SetGoalsService(goals *service.GoalsService) { h.goals = goals }

func NewTransactionsHandler(repo repository.TransactionRepository, classifier *service.ClassifierService) *TransactionsHandler {
	return &TransactionsHandler{
		repo:       repo,
		classifier: classifier,
	}
}

// ListTransactions lista as transações com filtros.
func (h *TransactionsHandler) ListTransactions(c echo.Context) error {
	userID := c.Get("user_id").(string)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize <= 0 {
		pageSize = 50
	} else if pageSize > 100 {
		pageSize = 100
	}
	search := strings.TrimSpace(c.QueryParam("q"))
	if search == "" {
		search = strings.TrimSpace(c.QueryParam("search"))
	}

	filters := repository.TransactionFilters{
		UserID:     userID,
		Search:     search,
		AccountID:  c.QueryParam("account_id"),
		CategoryID: c.QueryParam("category_id"),
		Direction:  c.QueryParam("direction"),
		Page:       page,
		PageSize:   pageSize,
	}
	if direction := c.QueryParam("direction"); direction != "" && direction != "debit" && direction != "credit" {
		return response.Error(c, http.StatusBadRequest, "tipo de movimentação inválido")
	}
	if rawReview := c.QueryParam("needs_review"); rawReview != "" {
		needsReview, err := strconv.ParseBool(rawReview)
		if err != nil {
			return response.Error(c, http.StatusBadRequest, "needs_review deve ser true ou false")
		}
		filters.NeedsReview = &needsReview
	}
	if rawIDs := c.QueryParam("ids"); rawIDs != "" {
		for _, id := range strings.Split(rawIDs, ",") {
			if id = strings.TrimSpace(id); id != "" && len(filters.IDs) < 50 {
				filters.IDs = append(filters.IDs, id)
			}
		}
	}

	if from := c.QueryParam("from_date"); from != "" {
		var err error
		filters.FromDate, err = time.Parse("2006-01-02", from)
		if err != nil {
			return response.Error(c, http.StatusBadRequest, "from_date inválida")
		}
	}
	if to := c.QueryParam("to_date"); to != "" {
		var err error
		filters.ToDate, err = time.Parse("2006-01-02", to)
		if err != nil {
			return response.Error(c, http.StatusBadRequest, "to_date inválida")
		}
	}

	transactions, total, err := h.repo.GetTransactions(c.Request().Context(), filters)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar transações")
	}

	totalPages := (total + pageSize - 1) / pageSize

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
		"total_pages":  totalPages,
	})
}

// UpdateCategory atualiza a categoria de uma transação.
func (h *TransactionsHandler) UpdateCategory(c echo.Context) error {
	userID := c.Get("user_id").(string)
	txID := c.Param("id")

	var req struct {
		CategoryID string `json:"category_id"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	if req.CategoryID == "" {
		return response.Error(c, http.StatusBadRequest, "a categoria é obrigatória")
	}

	updated, err := h.repo.UpdateCategory(c.Request().Context(), userID, txID, req.CategoryID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao atualizar categoria")
	}
	if !updated {
		return response.Error(c, http.StatusNotFound, "transação ou categoria não encontrada")
	}
	if h.goals != nil {
		if err := h.goals.UpdateGoalProgress(c.Request().Context(), userID); err != nil {
			return response.Error(c, http.StatusInternalServerError, "erro ao recalcular metas")
		}
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"message": "categoria atualizada com sucesso",
	})
}

// CreateManual registra uma entrada ou saída digitada pelo usuário.
func (h *TransactionsHandler) CreateManual(c echo.Context) error {
	userID := c.Get("user_id").(string)
	var req struct {
		AccountID   string  `json:"account_id"`
		Description string  `json:"description"`
		Direction   string  `json:"direction"`
		Date        string  `json:"date"`
		Amount      float64 `json:"amount"`
		CategoryID  string  `json:"category_id"`
	}
	if err := c.Bind(&req); err != nil || req.AccountID == "" || req.Description == "" || req.Amount <= 0 || (req.Direction != "credit" && req.Direction != "debit") {
		return response.Error(c, http.StatusBadRequest, "conta, descrição, valor e tipo válidos são obrigatórios")
	}
	date := time.Now()
	if req.Date != "" {
		var err error
		date, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			return response.Error(c, http.StatusBadRequest, "data inválida")
		}
	}
	id, err := h.repo.CreateManual(c.Request().Context(), userID, req.AccountID, req.Description, req.Direction, req.CategoryID, req.Amount, date)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "conta inválida")
	}
	if h.goals != nil {
		if err := h.goals.UpdateGoalProgress(c.Request().Context(), userID); err != nil {
			return response.Error(c, http.StatusInternalServerError, "erro ao recalcular metas")
		}
	}
	return response.Success(c, http.StatusCreated, map[string]string{"id": id})
}

// Summary retorna o resumo financeiro do período.
func (h *TransactionsHandler) Summary(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var fromDate, toDate time.Time
	var err error

	fromStr := c.QueryParam("from_date")
	toStr := c.QueryParam("to_date")

	if fromStr != "" && toStr != "" {
		fromDate, _ = time.Parse("2006-01-02", fromStr)
		toDate, _ = time.Parse("2006-01-02", toStr)
	} else if fromStr == "" && toStr == "" {
		// Se não passar nada, não filtra por data (período total)
		// No repository já tratamos o caso de IsZero()
	} else {
		// Comportamento padrão: mês atual
		now := time.Now()
		fromDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		toDate = fromDate.AddDate(0, 1, -1)
	}

	summary, err := h.repo.GetSummary(c.Request().Context(), userID, fromDate, toDate)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao gerar resumo financeiro")
	}

	return response.Success(c, http.StatusOK, summary)
}

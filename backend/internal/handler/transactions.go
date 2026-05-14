package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type TransactionsHandler struct {
	repo       repository.TransactionRepository
	classifier *service.ClassifierService
}

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
	}

	filters := repository.TransactionFilters{
		UserID:     userID,
		AccountID:  c.QueryParam("account_id"),
		CategoryID: c.QueryParam("category_id"),
		Direction:  c.QueryParam("direction"),
		Page:       page,
		PageSize:   pageSize,
	}

	if from := c.QueryParam("from_date"); from != "" {
		filters.FromDate, _ = time.Parse("2006-01-02", from)
	}
	if to := c.QueryParam("to_date"); to != "" {
		filters.ToDate, _ = time.Parse("2006-01-02", to)
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

	err := h.repo.UpdateCategory(c.Request().Context(), txID, req.CategoryID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao atualizar categoria")
	}

	return response.Success(c, http.StatusOK, map[string]string{
		"message": "categoria atualizada com sucesso",
	})
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

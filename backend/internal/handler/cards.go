package handler

import (
	"net/http"
	"time"

	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type CardsHandler struct {
	installmentService  *service.InstallmentsService
	subscriptionService *service.SubscriptionService
}

func NewCardsHandler(installmentService *service.InstallmentsService, subscriptionService *service.SubscriptionService) *CardsHandler {
	return &CardsHandler{
		installmentService:  installmentService,
		subscriptionService: subscriptionService,
	}
}

// ListInstallments lista todos os parcelamentos ativos.
func (h *CardsHandler) ListInstallments(c echo.Context) error {
	userID := c.Get("user_id").(string)

	installments, err := h.installmentService.GetActiveInstallments(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao listar parcelamentos")
	}

	return response.Success(c, http.StatusOK, installments)
}

// GetInvoice retorna a fatura projetada de um cartão.
func (h *CardsHandler) GetInvoice(c echo.Context) error {
	accountID := c.Param("account_id")

	monthParam := c.QueryParam("month") // YYYY-MM
	referenceMonth := time.Now()
	if monthParam != "" {
		parsed, err := time.Parse("2006-01", monthParam)
		if err == nil {
			referenceMonth = parsed
		}
	}

	invoice, err := h.installmentService.GetProjectedInvoice(c.Request().Context(), accountID, referenceMonth)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao calcular fatura projetada")
	}

	return response.Success(c, http.StatusOK, invoice)
}

// ListSubscriptions lista assinaturas recorrentes detectadas.
func (h *CardsHandler) ListSubscriptions(c echo.Context) error {
	userID := c.Get("user_id").(string)

	subscriptions, err := h.subscriptionService.DetectSubscriptions(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao detectar assinaturas")
	}

	return response.Success(c, http.StatusOK, subscriptions)
}

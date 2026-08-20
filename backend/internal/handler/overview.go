package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type OverviewHandler struct {
	db           *pgxpool.Pool
	transactions repository.TransactionRepository
	feed         *service.FeedService
	installments *service.InstallmentsService
}

type overviewCommitment struct {
	Kind    string    `json:"kind"`
	Title   string    `json:"title"`
	Amount  float64   `json:"amount"`
	DueDate time.Time `json:"due_date"`
}

func NewOverviewHandler(db *pgxpool.Pool, transactions repository.TransactionRepository, feed *service.FeedService, installments *service.InstallmentsService) *OverviewHandler {
	return &OverviewHandler{db: db, transactions: transactions, feed: feed, installments: installments}
}

func (h *OverviewHandler) Get(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Get("user_id").(string)
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao configurar horário financeiro")
	}
	now := time.Now().In(loc)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	summary, err := h.transactions.GetSummary(ctx, userID, monthStart, now)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao montar visão geral")
	}
	recent, _, err := h.transactions.GetTransactions(ctx, repository.TransactionFilters{UserID: userID, Page: 1, PageSize: 4})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar movimentações recentes")
	}

	needsReview, lastSynced, syncStatus, err := h.syncSummary(ctx, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao consultar sincronização")
	}
	events, err := h.feed.GetFeed(ctx, userID, 1, 20)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar alerta principal")
	}
	installments, err := h.installments.GetActiveInstallments(ctx, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar compromissos")
	}
	elapsedDays := float64(now.Day())
	weeklyProjected := summary.WeeklySpent
	if summary.TotalSpent > 0 {
		weeklyProjected = summary.TotalSpent / elapsedDays * 7
	}

	return response.Success(c, http.StatusOK, map[string]interface{}{
		"summary": summary,
		"balances": map[string]float64{
			"checking":  summary.CheckingBalance,
			"credit":    summary.CreditBalance,
			"available": summary.CheckingBalance - summary.ClosedInvoice - summary.MonthInstallments,
		},
		"weekly_pace": map[string]float64{
			"spent":          summary.WeeklySpent,
			"daily_average":  summary.WeeklySpent / 7,
			"projected_week": weeklyProjected,
		},
		"sync": map[string]interface{}{
			"status":     syncStatus,
			"updated_at": lastSynced,
		},
		"needs_review_count":  needsReview,
		"main_alert":          selectMainAlert(events),
		"next_commitment":     selectNextCommitment(installments, nil, now),
		"recent_transactions": recent,
	})
}

func (h *OverviewHandler) syncSummary(ctx context.Context, userID string) (int, *time.Time, string, error) {
	var needsReview int
	var lastSynced *time.Time
	err := h.db.QueryRow(ctx, `
		SELECT COUNT(t.id) FILTER (WHERE t.needs_review), MAX(a.last_synced_at)
		FROM connected_accounts a
		LEFT JOIN transactions t ON t.account_id = a.id
		WHERE a.user_id = $1
	`, userID).Scan(&needsReview, &lastSynced)
	if err != nil {
		return 0, nil, "", err
	}
	var status string
	err = h.db.QueryRow(ctx, `
		SELECT COALESCE((
			SELECT status FROM sync_logs WHERE user_id = $1 ORDER BY started_at DESC LIMIT 1
		), 'never')
	`, userID).Scan(&status)
	return needsReview, lastSynced, status, err
}

func selectMainAlert(events []service.FeedEvent) *service.FeedEvent {
	priorities := map[string]int{"info": 1, "warning": 2, "alert": 3}
	var selected *service.FeedEvent
	for i := range events {
		event := &events[i]
		if event.ReadAt != nil {
			continue
		}
		if selected == nil || priorities[event.Severity] > priorities[selected.Severity] {
			selected = event
		}
	}
	return selected
}

// Recorrências inferidas não são compromissos até serem confirmadas pelo Usuário.
func selectNextCommitment(installments []service.ActiveInstallment, _ []service.Subscription, now time.Time) *overviewCommitment {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var selected *overviewCommitment
	for _, installment := range installments {
		candidate := overviewCommitment{Kind: "installment", Title: installment.MerchantName, Amount: installment.Amount, DueDate: installment.NextDueDate}
		if candidate.DueDate.IsZero() || candidate.DueDate.Before(today) {
			continue
		}
		if selected == nil || candidate.DueDate.Before(selected.DueDate) {
			selected = &candidate
		}
	}
	return selected
}

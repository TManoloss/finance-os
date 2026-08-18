package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/finance-os/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

type categoryUpdateRepo struct {
	userID string
	allow  bool
}

func (r *categoryUpdateRepo) UpdateCategory(_ context.Context, userID, _, _ string) (bool, error) {
	r.userID = userID
	return r.allow, nil
}

func (*categoryUpdateRepo) GetTransactions(context.Context, repository.TransactionFilters) ([]map[string]interface{}, int, error) {
	return nil, 0, nil
}

func (*categoryUpdateRepo) GetSummary(context.Context, string, time.Time, time.Time) (*repository.TransactionSummary, error) {
	return nil, nil
}

func (*categoryUpdateRepo) CreateManual(context.Context, string, string, string, string, string, float64, time.Time) (string, error) {
	return "", nil
}

func TestUpdateCategoryUsesAuthenticatedUser(t *testing.T) {
	repo := &categoryUpdateRepo{}
	handler := NewTransactionsHandler(repo, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/transactions/tx/category", strings.NewReader(`{"category_id":"category"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "owner")
	c.SetParamNames("id")
	c.SetParamValues("tx")

	if err := handler.UpdateCategory(c); err != nil {
		t.Fatal(err)
	}
	if repo.userID != "owner" {
		t.Fatalf("expected authenticated user, got %q", repo.userID)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an unowned transaction/category, got %d", rec.Code)
	}
}

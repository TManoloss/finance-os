package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/finance-os/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

type categoryUpdateRepo struct {
	userID  string
	allow   bool
	calls   int
	filters repository.TransactionFilters
	total   int
}

func (r *categoryUpdateRepo) UpdateCategory(_ context.Context, userID, _, _ string) (bool, error) {
	r.userID = userID
	return r.allow, nil
}

func (r *categoryUpdateRepo) GetTransactions(_ context.Context, filters repository.TransactionFilters) ([]map[string]interface{}, int, error) {
	r.calls++
	r.filters = filters
	return nil, r.total, nil
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

func TestListTransactionsReadsRelatedIDs(t *testing.T) {
	repo := &categoryUpdateRepo{}
	handler := NewTransactionsHandler(repo, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/transactions?ids=tx-1,tx-2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "owner")

	if err := handler.ListTransactions(c); err != nil {
		t.Fatal(err)
	}
	if strings.Join(repo.filters.IDs, ",") != "tx-1,tx-2" {
		t.Fatalf("ids relacionados = %#v", repo.filters.IDs)
	}
}

func TestListTransactionsParsesFilters(t *testing.T) {
	for _, wantNeedsReview := range []bool{true, false} {
		t.Run(strconv.FormatBool(wantNeedsReview), func(t *testing.T) {
			repo := &categoryUpdateRepo{total: 51}
			handler := NewTransactionsHandler(repo, nil)
			e := echo.New()
			query := url.Values{
				"q":            {"mercado livre"},
				"needs_review": {strconv.FormatBool(wantNeedsReview)},
				"account_id":   {"account-1"},
				"category_id":  {"category-1"},
				"direction":    {"debit"},
				"from_date":    {"2026-08-01"},
				"to_date":      {"2026-08-20"},
				"page":         {"3"},
				"page_size":    {"25"},
			}
			req := httptest.NewRequest(http.MethodGet, "/transactions?"+query.Encode(), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "owner")

			if err := handler.ListTransactions(c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}

			if repo.calls != 1 {
				t.Fatalf("repositório chamado %d vezes, quer 1", repo.calls)
			}
			if repo.filters.UserID != "owner" {
				t.Fatalf("user_id = %q", repo.filters.UserID)
			}
			if repo.filters.Search != "mercado livre" {
				t.Fatalf("q = %q", repo.filters.Search)
			}
			if repo.filters.NeedsReview == nil || *repo.filters.NeedsReview != wantNeedsReview {
				t.Fatalf("needs_review = %v, quer %t", repo.filters.NeedsReview, wantNeedsReview)
			}
			if repo.filters.AccountID != "account-1" {
				t.Fatalf("account_id = %q", repo.filters.AccountID)
			}
			if repo.filters.CategoryID != "category-1" {
				t.Fatalf("category_id = %q", repo.filters.CategoryID)
			}
			if repo.filters.Direction != "debit" {
				t.Fatalf("direction = %q", repo.filters.Direction)
			}
			if got, want := repo.filters.FromDate, time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
				t.Fatalf("from_date = %v, quer %v", got, want)
			}
			if got, want := repo.filters.ToDate, time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
				t.Fatalf("to_date = %v, quer %v", got, want)
			}
			if repo.filters.Page != 3 || repo.filters.PageSize != 25 {
				t.Fatalf("paginação = page %d/page_size %d, quer 3/25", repo.filters.Page, repo.filters.PageSize)
			}

			var response struct {
				Data struct {
					Total      int `json:"total"`
					Page       int `json:"page"`
					PageSize   int `json:"page_size"`
					TotalPages int `json:"total_pages"`
				} `json:"data"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if response.Data.Total != 51 || response.Data.Page != 3 || response.Data.PageSize != 25 || response.Data.TotalPages != 3 {
				t.Fatalf("metadados de paginação = %+v", response.Data)
			}
		})
	}
}

func TestListTransactionsRejectsInvalidNeedsReview(t *testing.T) {
	repo := &categoryUpdateRepo{}
	handler := NewTransactionsHandler(repo, nil)
	e := echo.New()
	query := url.Values{"needs_review": {"maybe"}}
	req := httptest.NewRequest(http.MethodGet, "/transactions?"+query.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "owner")

	if err := handler.ListTransactions(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if repo.calls != 0 {
		t.Fatalf("repositório chamado %d vezes para needs_review inválido", repo.calls)
	}
}

package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateGoalRejectsInvalidDatesBeforeServiceAccess(t *testing.T) {
	for name, body := range map[string]string{
		"start date":  `{"name":"Reserva","goal_type":"savings","target_amount":100,"start_date":"2026-02-30"}`,
		"target date": `{"name":"Reserva","goal_type":"savings","target_amount":100,"start_date":"2026-08-20","target_date":"2026-02-30"}`,
		"date order":  `{"name":"Reserva","goal_type":"savings","target_amount":100,"start_date":"2026-08-20","target_date":"2026-08-19"}`,
	} {
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/goals", strings.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", "user-1")

			if err := NewGoalsHandler(nil, nil).Create(c); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
		})
	}
}

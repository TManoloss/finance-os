package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VisualReportsService struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewVisualReportsService(db *pgxpool.Pool, cfg *config.Config) *VisualReportsService {
	return &VisualReportsService{db: db, cfg: cfg}
}

// GetSpendingHeatmap returns daily spending totals for the last 365 days
func (s *VisualReportsService) GetSpendingHeatmap(userID string) (interface{}, error) {
	query := `
		SELECT t.date, SUM(t.amount) as total
		FROM transactions t
		JOIN connected_accounts a ON t.account_id = a.id
		WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= CURRENT_DATE - INTERVAL '365 days'
		GROUP BY t.date
		ORDER BY t.date ASC
	`
	
	rows, err := s.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type HeatmapDay struct {
		Date  string  `json:"date"`
		Total float64 `json:"total"`
	}

	var heatmap []HeatmapDay
	for rows.Next() {
		var date time.Time
		var total float64
		if err := rows.Scan(&date, &total); err != nil {
			return nil, err
		}
		heatmap = append(heatmap, HeatmapDay{
			Date:  date.Format("2006-01-02"),
			Total: total,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return heatmap, nil
}

func (s *VisualReportsService) GetDependencyMap(userID string) (interface{}, error) {
	url := fmt.Sprintf("%s/reports/dependency-map/%s", s.cfg.AgentsServiceURL, userID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *VisualReportsService) GetMonthlyReplay(userID, month string) (interface{}, error) {
	url := fmt.Sprintf("%s/reports/monthly-replay/%s?month=%s", s.cfg.AgentsServiceURL, userID, month)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

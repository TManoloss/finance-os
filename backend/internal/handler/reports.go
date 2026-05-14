package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/response"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type ReportsHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewReportsHandler(db *pgxpool.Pool, cfg *config.Config) *ReportsHandler {
	return &ReportsHandler{
		db:  db,
		cfg: cfg,
	}
}

// TriggerAgent dispara manualmente a geração de um relatório.
func (h *ReportsHandler) TriggerAgent(c echo.Context) error {
	userID := c.Get("user_id").(string)
	agentType := c.Param("type") // daily, weekly, monthly

	if agentType != "daily" && agentType != "weekly" && agentType != "monthly" {
		return response.Error(c, http.StatusBadRequest, "tipo de agente inválido")
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/agents/%s/%s", h.cfg.AgentsServiceURL, agentType, userID)
	
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	return response.Success(c, http.StatusAccepted, map[string]string{
		"message": fmt.Sprintf("processamento do agente %s solicitado", agentType),
	})
}

// GetReports retorna a lista de relatórios gerados pelos agentes.
func (h *ReportsHandler) GetReports(c echo.Context) error {
	userID := c.Get("user_id").(string)

	rows, err := h.db.Query(c.Request().Context(), `
		SELECT id, agent_type, period_start, period_end, summary_markdown, insights, created_at 
		FROM agent_reports WHERE user_id = $1 ORDER BY created_at DESC LIMIT 10`, userID)

	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar relatórios")
	}
	defer rows.Close()

	var reports []map[string]interface{}
	for rows.Next() {
		var r struct {
			ID              string
			AgentType       string
			PeriodStart     interface{}
			PeriodEnd       interface{}
			SummaryMarkdown string
			Insights        interface{}
			CreatedAt       interface{}
		}
		err := rows.Scan(&r.ID, &r.AgentType, &r.PeriodStart, &r.PeriodEnd, &r.SummaryMarkdown, &r.Insights, &r.CreatedAt)
		if err != nil {
			continue
		}
		reports = append(reports, map[string]interface{}{
			"id":               r.ID,
			"agent_type":       r.AgentType,
			"period_start":     r.PeriodStart,
			"period_end":       r.PeriodEnd,
			"summary_markdown": r.SummaryMarkdown,
			"insights":         r.Insights,
			"created_at":       r.CreatedAt,
		})
	}

	return response.Success(c, http.StatusOK, reports)
}

// GetCashflow retorna o fluxo de caixa diário e padrões detectados.
func (h *ReportsHandler) GetCashflow(c echo.Context) error {
	userID := c.Get("user_id").(string)
	from := c.QueryParam("from")
	to := c.QueryParam("to")

	if from == "" || to == "" {
		return response.Error(c, http.StatusBadRequest, "parâmetros 'from' e 'to' são obrigatórios")
	}

	url := fmt.Sprintf("%s/reports/cashflow/%s?from_date=%s&to_date=%s", h.cfg.AgentsServiceURL, userID, from, to)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response.Error(c, resp.StatusCode, "erro retornado pelo serviço de agentes")
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetBehavioralInsights retorna insights comportamentais gerados por IA.
func (h *ReportsHandler) GetBehavioralInsights(c echo.Context) error {
	userID := c.Get("user_id").(string)

	url := fmt.Sprintf("%s/reports/behavioral/%s", h.cfg.AgentsServiceURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetInvisibleSpending busca gastos invisíveis (assinaturas esquecidas, duplicatas, etc).
func (h *ReportsHandler) GetInvisibleSpending(c echo.Context) error {
	userID := c.Get("user_id").(string)

	url := fmt.Sprintf("%s/reports/invisible-spending/%s", h.cfg.AgentsServiceURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetProjections retorna as projeções financeiras para o fim do mês e próximos 3 meses.
func (h *ReportsHandler) GetProjections(c echo.Context) error {
	userID := c.Get("user_id").(string)

	url := fmt.Sprintf("%s/reports/projection/%s", h.cfg.AgentsServiceURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetHealthScore retorna o score de saúde financeira e recomendações.
func (h *ReportsHandler) GetHealthScore(c echo.Context) error {
	userID := c.Get("user_id").(string)

	url := fmt.Sprintf("%s/reports/health-score/%s", h.cfg.AgentsServiceURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetTopMerchants retorna a lista de principais estabelecimentos.
func (h *ReportsHandler) GetTopMerchants(c echo.Context) error {
	userID := c.Get("user_id").(string)
	months := c.QueryParam("months")
	if months == "" {
		months = "3"
	}

	url := fmt.Sprintf("%s/merchants/%s?months=%s", h.cfg.AgentsServiceURL, userID, months)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetMerchantProfile retorna o perfil detalhado de um estabelecimento.
func (h *ReportsHandler) GetMerchantProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)
	merchantName := c.Param("name")

	url := fmt.Sprintf("%s/merchants/%s/%s", h.cfg.AgentsServiceURL, userID, merchantName)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetUpcomingExpenses retorna previsões de despesas futuras (sazonalidade).
func (h *ReportsHandler) GetUpcomingExpenses(c echo.Context) error {
	userID := c.Get("user_id").(string)

	url := fmt.Sprintf("%s/reports/upcoming-expenses/%s", h.cfg.AgentsServiceURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}

// GetNarrativeReport retorna o relatório mensal em formato de narrativa IA.
func (h *ReportsHandler) GetNarrativeReport(c echo.Context) error {
	userID := c.Get("user_id").(string)
	month := c.QueryParam("month")
	year := c.QueryParam("year")

	if month == "" || year == "" {
		return response.Error(c, http.StatusBadRequest, "parâmetros 'month' e 'year' são obrigatórios")
	}

	url := fmt.Sprintf("%s/reports/narrative/%s?month=%s&year=%s", h.cfg.AgentsServiceURL, userID, month, year)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}
func (h *ReportsHandler) GetComparison(c echo.Context) error {
	userID := c.Get("user_id").(string)
	aStart := c.QueryParam("a_start")
	aEnd := c.QueryParam("a_end")
	bStart := c.QueryParam("b_start")
	bEnd := c.QueryParam("b_end")

	if aStart == "" || aEnd == "" || bStart == "" || bEnd == "" {
		return response.Error(c, http.StatusBadRequest, "todos os parâmetros de período (a_start, a_end, b_start, b_end) são obrigatórios")
	}

	url := fmt.Sprintf("%s/reports/comparison/%s?a_start=%s&a_end=%s&b_start=%s&b_end=%s", 
		h.cfg.AgentsServiceURL, userID, aStart, aEnd, bStart, bEnd)

	resp, err := http.Get(url)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "falha ao comunicar com o serviço de agentes")
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao decodificar resposta")
	}

	return response.Success(c, http.StatusOK, result)
}


package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type ChatHandler struct {
	cfg                  *config.Config
	healthService        *service.FinancialHealthService
	visualReportsService *service.VisualReportsService
}

func NewChatHandler(cfg *config.Config, healthService *service.FinancialHealthService, visualReportsService *service.VisualReportsService) *ChatHandler {
	return &ChatHandler{cfg: cfg, healthService: healthService, visualReportsService: visualReportsService}
}

// SendMessage envia uma mensagem para o agente de chat IA.
func (h *ChatHandler) SendMessage(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		Message string              `json:"message" validate:"required"`
		History []map[string]string `json:"history"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}
	if strings.TrimSpace(req.Message) == "" {
		return response.Error(c, http.StatusBadRequest, "a mensagem é obrigatória")
	}

	ctx := c.Request().Context()
	calculatedContext := make(map[string]interface{}, 2)

	intelligence, intelligenceErr := h.healthService.GetConsolidatedIntelligence(ctx, userID)
	if intelligenceErr != nil {
		log.Printf("[Chat] Falha ao carregar inteligência calculada: %v", intelligenceErr)
	} else {
		calculatedContext["intelligence"] = intelligence
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao configurar horário financeiro")
	}
	replay, replayErr := h.visualReportsService.GetMonthlyReplay(ctx, userID, time.Now().In(loc).Format("2006-01"))
	if replayErr != nil {
		log.Printf("[Chat] Falha ao carregar replay calculado: %v", replayErr)
	} else {
		calculatedContext["monthly_replay"] = replay
	}

	if len(calculatedContext) == 0 {
		return response.Error(c, http.StatusServiceUnavailable, "dados calculados indisponíveis no momento")
	}

	payload := map[string]interface{}{
		"user_id": userID,
		"message": req.Message,
		"history": req.History,
		"context": calculatedContext,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao preparar contexto calculado")
	}
	fallback := deterministicChatFallback(intelligence, replay)

	agentRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, h.cfg.AgentsServiceURL+"/chat", bytes.NewBuffer(body))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao preparar assistente")
	}
	agentRequest.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(agentRequest)
	if err != nil {
		log.Printf("[Chat] Erro ao chamar serviço Python: %v", err)
		return response.Success(c, http.StatusOK, map[string]interface{}{"response": fallback, "fallback": true})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		log.Printf("[Chat] Serviço Python retornou erro (%d): %v", resp.StatusCode, errBody)
		return response.Success(c, http.StatusOK, map[string]interface{}{"response": fallback, "fallback": true})
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[Chat] Erro ao decodificar resposta do agente: %v", err)
		return response.Success(c, http.StatusOK, map[string]interface{}{"response": fallback, "fallback": true})
	}
	if answer, ok := result["response"].(string); !ok || strings.TrimSpace(answer) == "" {
		return response.Success(c, http.StatusOK, map[string]interface{}{"response": fallback, "fallback": true})
	}

	log.Printf("[Chat] Resposta recebida do agente")
	return response.Success(c, http.StatusOK, result)
}

func deterministicChatFallback(intelligence *service.ConsolidatedIntelligence, replay *service.MonthlyReplayResult) string {
	parts := []string{"A explicação por IA está indisponível agora, mas os cálculos do FinanceOS continuam funcionando."}
	if intelligence != nil {
		status := map[string]string{"excellent": "excelente", "good": "boa", "fair": "regular", "attention": "em atenção", "critical": "crítica"}[intelligence.HealthStatus]
		if status == "" {
			status = "sem classificação"
		}
		quality := map[string]string{"high": "alta", "medium": "média", "low": "baixa"}[intelligence.Quality]
		if quality == "" {
			quality = "não informada"
		}
		parts = append(parts, fmt.Sprintf("Sua saúde financeira está em **%.0f/100** (%s), com qualidade de dados %s.", intelligence.OverallHealthScore, status, quality))
	}
	if replay != nil {
		for _, insight := range replay.Insights {
			parts = append(parts, insight)
		}
		if replay.NextMonthGuidance != "" {
			parts = append(parts, replay.NextMonthGuidance)
		}
	}
	return strings.Join(parts, "\n\n")
}

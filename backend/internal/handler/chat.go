package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/response"
	"github.com/labstack/echo/v4"
)

type ChatHandler struct {
	cfg *config.Config
}

func NewChatHandler(cfg *config.Config) *ChatHandler {
	return &ChatHandler{cfg: cfg}
}

// SendMessage envia uma mensagem para o agente de chat IA.
func (h *ChatHandler) SendMessage(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req struct {
		Message string `json:"message" validate:"required"`
		History []map[string]string `json:"history"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	// Encaminha para o serviço Python
	payload := map[string]interface{}{
		"user_id": userID,
		"message": req.Message,
		"history": req.History,
	}
	body, _ := json.Marshal(payload)

	log.Printf("[Chat] Enviando mensagem para agente: %s", req.Message)
	resp, err := http.Post(h.cfg.AgentsServiceURL+"/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[Chat] Erro ao chamar serviço Python: %v", err)
		return response.Error(c, http.StatusInternalServerError, "erro ao contatar assistente")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		log.Printf("[Chat] Serviço Python retornou erro (%d): %v", resp.StatusCode, errBody)
		return response.Error(c, http.StatusInternalServerError, "assistente indisponível no momento")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[Chat] Erro ao decodificar resposta do agente: %v", err)
		return response.Error(c, http.StatusInternalServerError, "erro no processamento da resposta")
	}

	log.Printf("[Chat] Resposta recebida do agente")
	return response.Success(c, http.StatusOK, result)
}


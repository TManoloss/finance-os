package pluggy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL = "https://api.pluggy.ai"
)

// Client é o cliente para interação com a API da Pluggy.
type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	apiKey       string
	mu           sync.RWMutex
	expiresAt    time.Time
}

// NewClient cria uma nova instância do cliente Pluggy.
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate realiza a autenticação na Pluggy e armazena a API Key.
// A chave é válida por 2 horas, implementamos um cache em memória.
func (c *Client) Authenticate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Se já temos uma chave válida por mais 5 minutos, não autentica de novo
	if c.apiKey != "" && time.Now().Add(5*time.Minute).Before(c.expiresAt) {
		return nil
	}

	payload := AuthRequest{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	}

	// Log para debug de formato (mascarado)
	idLen := len(payload.ClientID)
	maskedID := ""
	if idLen > 8 {
		maskedID = payload.ClientID[:4] + "..." + payload.ClientID[idLen-4:]
	}
	log.Printf("[Pluggy] Tentando autenticar com ClientID: %s (tamanho: %d)", maskedID, idLen)

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(baseURL+"/auth", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[Pluggy] Erro de rede na autenticação: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errRes map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errRes)
		log.Printf("[Pluggy] Falha na autenticação (Status %d): %v", resp.StatusCode, errRes)
		return fmt.Errorf("falha na autenticação pluggy: status %d", resp.StatusCode)
	}

	var authRes AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authRes); err != nil {
		return err
	}

	c.apiKey = authRes.APIKey
	c.expiresAt = time.Now().Add(2 * time.Hour)

	return nil
}

// doRequest é um helper para fazer requisições autenticadas.
func (c *Client) doRequest(method, path string, body []byte) (*http.Response, error) {
	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	apiKey := c.apiKey
	c.mu.RUnlock()

	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// CreateConnectToken gera um token de acesso para o widget Pluggy Connect.
// Se itemID não for nulo, o token será para atualizar uma conexão existente.
func (c *Client) CreateConnectToken(itemID *string) (string, error) {
	payload := make(map[string]interface{})
	if itemID != nil && *itemID != "" {
		payload["itemId"] = *itemID
	}

	body, _ := json.Marshal(payload)
	resp, err := c.doRequest(http.MethodPost, "/connect_token", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errRes map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errRes)
		return "", fmt.Errorf("erro ao gerar connect token: status %d, %v", resp.StatusCode, errRes)
	}

	var res ConnectTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

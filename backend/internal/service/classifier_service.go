package service

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/finance-os/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClassifierService lida com a lógica de classificação de transações.
type ClassifierService struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

// NewClassifierService cria uma nova instância de ClassifierService.
func NewClassifierService(db *pgxpool.Pool, cfg *config.Config) *ClassifierService {
	return &ClassifierService{
		db:  db,
		cfg: cfg,
	}
}

func (s *ClassifierService) cleanText(text string) string {
	// Remove prefixos de Pix/Transferência
	re := regexp.MustCompile(`(?i)^(TRANSFER[ÊE]NCIA|PIX|TED|DOC)\s+(ENVIAD[OA]|RECEBID[OA])\s*[|:]\s*`)
	text = re.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

// Classify tenta encontrar a melhor categoria para uma transação.
func (s *ClassifierService) Classify(ctx context.Context, userID, merchantName, description string, amount float64, direction string) (string, error) {
	merchantClean := s.cleanText(merchantName)
	descriptionClean := s.cleanText(description)

	// 1. Verificar regra exata por merchant_name (limpo)
	var categoryID string
	err := s.db.QueryRow(ctx, `
		SELECT category_id FROM category_rules 
		WHERE (user_id = $1 OR user_id IS NULL) 
		AND LOWER(merchant_pattern) = LOWER($2)
		ORDER BY user_id NULLS LAST, priority DESC
		LIMIT 1`, userID, merchantClean).Scan(&categoryID)

	if err == nil {
		return categoryID, nil
	}

	// 2. Verificar regra por padrão LIKE na descrição (limpa)
	err = s.db.QueryRow(ctx, `
		SELECT category_id FROM category_rules 
		WHERE (user_id = $1 OR user_id IS NULL) 
		AND $2 ILIKE '%' || merchant_pattern || '%'
		ORDER BY user_id NULLS LAST, priority DESC
		LIMIT 1`, userID, descriptionClean).Scan(&categoryID)

	if err == nil {
		return categoryID, nil
	}

	// 3. Chamar serviço Python de IA
	log.Printf("[Classifier] Chamando IA para: %s / %s", merchantClean, descriptionClean)
	
	categoryID, err = s.classifyWithIA(ctx, merchantClean, descriptionClean, amount, direction)
	if err == nil && categoryID != "" {
		var realID string
		s.db.QueryRow(ctx, "SELECT id FROM categories WHERE name ILIKE $1 AND user_id IS NULL", categoryID).Scan(&realID)
		if realID != "" {
			return realID, nil
		}
	}

	// Fallback final
	var outrosID string
	s.db.QueryRow(ctx, "SELECT id FROM categories WHERE name = 'Outros' AND user_id IS NULL").Scan(&outrosID)

	return outrosID, nil
}

type IAClassifyRequest struct {
	MerchantName string  `json:"merchant_name"`
	Description  string  `json:"description"`
	Amount       float64 `json:"amount"`
	Direction    string  `json:"direction"`
}

func (s *ClassifierService) classifyWithIA(ctx context.Context, merchant, desc string, amount float64, direction string) (string, error) {
	payload := IAClassifyRequest{
		MerchantName: merchant,
		Description:  desc,
		Amount:       amount,
		Direction:    direction,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(s.cfg.AgentsServiceURL+"/classify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		CategoryID string `json:"category_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.CategoryID, nil
}

// CreateRule cria uma nova regra de classificação para o usuário.
func (s *ClassifierService) CreateRule(ctx context.Context, userID, merchantPattern, categoryID string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO category_rules (user_id, merchant_pattern, category_id, priority)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (user_id, merchant_pattern) DO UPDATE SET category_id = EXCLUDED.category_id`,
		userID, merchantPattern, categoryID)
	return err
}

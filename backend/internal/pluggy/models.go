package pluggy

import (
	"time"
)

// AuthRequest representa o corpo da requisição de autenticação.
type AuthRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// AuthResponse representa a resposta de sucesso na autenticação.
type AuthResponse struct {
	APIKey string `json:"apiKey"`
}

// Item representa uma conexão com uma instituição financeira.
type Item struct {
	ID          string    `json:"id"`
	ConnectorID int       `json:"connectorId"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Account representa uma conta bancária vinculada a um Item.
type Account struct {
	ID            string  `json:"id"`
	ItemID        string  `json:"itemId"`
	Type          string  `json:"type"`          // CHECKING, SAVINGS, CREDIT
	Subtype       string  `json:"subtype"`       // CHECKING_ACCOUNT, SAVINGS_ACCOUNT, CREDIT_CARD
	Name          string  `json:"name"`          // Nome personalizado da conta
	MarketingName string  `json:"marketingName"` // Nome exibido pelo banco (ex: "Conta Corrente")
	Balance       float64 `json:"balance"`
	CurrencyCode  string  `json:"currencyCode"`
	Number        string  `json:"number"`
}

// Transaction representa uma movimentação financeira em uma conta.
type Transaction struct {
	ID                 string              `json:"id"`
	AccountID          string              `json:"accountId"`
	Description        string              `json:"description"`
	Amount             float64             `json:"amount"`
	CurrencyCode       string              `json:"currencyCode"`
	Date               time.Time           `json:"date"`
	Category           string              `json:"category"`
	Type               string              `json:"type"` // DEBIT, CREDIT
	Status             string              `json:"status"`
	CreditCardMetadata *CreditCardMetadata `json:"creditCardMetadata"`
}

type CreditCardMetadata struct {
	InstallmentNumber int `json:"installmentNumber"`
	InstallmentsCount  int `json:"installmentsCount"`
}

// Connector representa uma instituição financeira suportada pela Pluggy.
type Connector struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PrimaryColor string `json:"primaryColor"`
	ImageUrl     string `json:"imageUrl"`
	Type         string `json:"type"`
}

// ConnectTokenResponse representa a resposta da geração de token para o widget.
type ConnectTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

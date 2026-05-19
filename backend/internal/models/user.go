package models

import (
	"time"
)

// User representa um usuário no sistema.
type User struct {
	ID                         string    `json:"id"`
	Name                       string    `json:"name"`
	Email                      string    `json:"email"`
	PasswordHash               string    `json:"-"`
	PluggyClientID           string `json:"pluggy_client_id"`
	PluggyClientSecretEncrypted string `json:"-"`
	GroqAPIKeyEncrypted        string `json:"-"`
	GeminiAPIKeyEncrypted      string `json:"-"`
	HasGroqKey                 bool   `json:"has_groq_key"`
	HasGeminiKey               bool   `json:"has_gemini_key"`
	CreatedAt                time.Time `json:"created_at"`
	}


// RefreshToken representa um token de atualização de sessão.
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

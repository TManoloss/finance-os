package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/finance-os/backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims define os claims personalizados do JWT.
type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService lida com a geração e validação de tokens JWT.
type JWTService struct {
	cfg *config.Config
}

// NewJWTService cria uma nova instância de JWTService.
func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{cfg: cfg}
}

// GenerateAccessToken gera um novo token de acesso (JWT) válido por 15 minutos.
func (s *JWTService) GenerateAccessToken(userID, email string) (string, error) {
	claims := &CustomClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "finance-os",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// GenerateRefreshToken gera um novo token de atualização opaco.
func (s *JWTService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidateAccessToken valida um token de acesso e retorna seus claims.
func (s *JWTService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inesperado")
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}

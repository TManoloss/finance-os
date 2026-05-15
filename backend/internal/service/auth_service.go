package service

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/finance-os/backend/internal/models"
	"github.com/finance-os/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("usuário não encontrado")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrEmailTaken        = errors.New("este email já está em uso")
	ErrInvalidPassword   = errors.New("a senha não atende aos requisitos de segurança")
)

// AuthService lida com a lógica de negócio de autenticação.
type AuthService struct {
	repo       repository.UserRepository
	jwtService *JWTService
}

// NewAuthService cria uma nova instância de AuthService.
func NewAuthService(repo repository.UserRepository, jwtService *JWTService) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtService: jwtService,
	}
}

// AuthResponse representa a resposta de sucesso na autenticação.
type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

// Register realiza o cadastro de um novo usuário.
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*AuthResponse, error) {
	// Normalizar email
	email = strings.ToLower(strings.TrimSpace(email))

	// 1. Validar política de senha
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// 2. Verificar se o e-mail já está em uso
	existingUser, _ := s.repo.FindByEmail(ctx, email)
	if existingUser != nil {
		return nil, ErrEmailTaken
	}

	// 3. Gerar hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Printf("[AuthService] Erro ao gerar hash: %v", err)
		return nil, err
	}
	log.Printf("[AuthService] Hash gerado para %s: %s", email, string(hashedPassword))

	// 4. Criar usuário no banco
	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 5. Gerar par de tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 6. Salvar Refresh Token
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 dias
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login realiza a autenticação de um usuário.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	// Normalizar email
	email = strings.ToLower(strings.TrimSpace(email))

	// 1. Buscar usuário por email
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		log.Printf("[AuthService] Erro ao buscar usuário %s: %v", email, err)
		return nil, ErrInvalidCredentials
	}

	// 2. Verificar senha
	log.Printf("[AuthService] Comparando senha para %s. Hash no banco: %s", email, user.PasswordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Printf("[AuthService] Falha na comparação de senha para %s: %v", email, err)
		return nil, ErrInvalidCredentials
	}

	// 3. Gerar par de tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 4. Salvar Refresh Token
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 dias
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GetUserByID busca os dados de um usuário pelo ID.
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// RefreshToken valida um refresh token e gera um novo par de tokens (rotação).
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// 1. Validar refresh token no banco
	userID, err := s.repo.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 2. Buscar dados do usuário
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 3. Deletar o token usado (rotação)
	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	// 4. Gerar novo par de tokens
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 5. Salvar o novo Refresh Token
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 dias
	if err := s.repo.SaveRefreshToken(ctx, user.ID, newRefreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// validatePassword verifica se a senha atende aos requisitos mínimos.
func (s *AuthService) validatePassword(password string) error {
	if len(password) < 10 {
		return ErrInvalidPassword
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Verifica caracteres especiais adicionais se IsPunct/IsSymbol não pegar tudo
	if !hasSpecial {
		specialCharMatch, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
		hasSpecial = specialCharMatch
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrInvalidPassword
	}

	return nil
}

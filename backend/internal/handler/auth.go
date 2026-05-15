package handler

import (
	"net/http"

	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// RegisterRequest representa o corpo da requisição de registro.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginRequest representa o corpo da requisição de login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest representa o corpo da requisição de refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Register trata o registro de novos usuários.
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return response.Error(c, http.StatusBadRequest, "todos os campos são obrigatórios")
	}

	res, err := h.authService.Register(c.Request().Context(), req.Name, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrEmailTaken:
			return response.Error(c, http.StatusConflict, err.Error())
		case service.ErrInvalidPassword:
			return response.Error(c, http.StatusBadRequest, err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "erro ao registrar usuário")
		}
	}

	return response.Success(c, http.StatusCreated, res)
}

// Login trata a autenticação de usuários.
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	if req.Email == "" || req.Password == "" {
		return response.Error(c, http.StatusBadRequest, "todos os campos são obrigatórios")
	}

	res, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			return response.Error(c, http.StatusUnauthorized, err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "erro ao realizar login")
		}
	}

	return response.Success(c, http.StatusOK, res)
}

// Refresh trata a renovação de tokens JWT.
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "formato de requisição inválido")
	}

	if req.RefreshToken == "" {
		return response.Error(c, http.StatusBadRequest, "o refresh_token é obrigatório")
	}

	res, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			return response.Error(c, http.StatusUnauthorized, "token de renovação inválido ou expirado")
		case service.ErrUserNotFound:
			return response.Error(c, http.StatusNotFound, err.Error())
		default:
			return response.Error(c, http.StatusInternalServerError, "erro ao renovar token")
		}
	}

	return response.Success(c, http.StatusOK, res)
}

// Me retorna os dados do usuário logado.
func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return response.Error(c, http.StatusUnauthorized, "usuário não autenticado")
	}

	user, err := h.authService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "usuário não encontrado")
	}

	return response.Success(c, http.StatusOK, user)
}

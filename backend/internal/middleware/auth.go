package middleware

import (
	"net/http"
	"strings"

	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware lida com a proteção de rotas usando JWT.
func AuthMiddleware(jwtService *service.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Error(c, http.StatusUnauthorized, "token de autenticação ausente")
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return response.Error(c, http.StatusUnauthorized, "formato de token inválido")
			}

			tokenString := parts[1]
			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				return response.Error(c, http.StatusUnauthorized, "token inválido ou expirado")
			}

			// Armazena os claims no contexto do Echo
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}

package response

import (
	"github.com/labstack/echo/v4"
)

// APIResponse é a estrutura padrão para todas as respostas da API.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success envia uma resposta de sucesso padronizada.
func Success(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error envia uma resposta de erro padronizada.
func Error(c echo.Context, status int, message string) error {
	return c.JSON(status, APIResponse{
		Success: false,
		Error:   message,
	})
}

package handler

import (
	"net/http"

	"github.com/finance-os/backend/internal/response"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type CategoriesHandler struct {
	db *pgxpool.Pool
}

func NewCategoriesHandler(db *pgxpool.Pool) *CategoriesHandler {
	return &CategoriesHandler{db: db}
}

// ListCategories lista todas as categorias do sistema e as customizadas do usuário.
func (h *CategoriesHandler) ListCategories(c echo.Context) error {
	userID := c.Get("user_id").(string)

	query := `
		SELECT id, name, color, icon, user_id
		FROM categories
		WHERE user_id = $1 OR user_id IS NULL
		ORDER BY name ASC
	`
	rows, err := h.db.Query(c.Request().Context(), query, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar categorias")
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var cat struct {
			ID     string
			Name   string
			Color  *string
			Icon   *string
			UserID *string
		}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Color, &cat.Icon, &cat.UserID); err != nil {
			continue
		}

		categories = append(categories, map[string]interface{}{
			"id":      cat.ID,
			"name":    cat.Name,
			"color":   cat.Color,
			"icon":    cat.Icon,
			"is_system": cat.UserID == nil,
		})
	}

	return response.Success(c, http.StatusOK, categories)
}

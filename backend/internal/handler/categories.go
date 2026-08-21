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
		SELECT DISTINCT ON (LOWER(c.name), COALESCE(c.parent_id::text, '')) c.id, c.name, c.color, c.icon, c.user_id, c.parent_id, parent.name
		FROM categories c
		LEFT JOIN categories parent ON parent.id = c.parent_id
		WHERE c.user_id = $1 OR c.user_id IS NULL
		ORDER BY LOWER(c.name), COALESCE(c.parent_id::text, ''), (c.user_id IS NOT NULL) DESC, c.created_at DESC
	`
	rows, err := h.db.Query(c.Request().Context(), query, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar categorias")
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var cat struct {
			ID         string
			Name       string
			Color      *string
			Icon       *string
			UserID     *string
			ParentID   *string
			ParentName *string
		}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Color, &cat.Icon, &cat.UserID, &cat.ParentID, &cat.ParentName); err != nil {
			continue
		}

		categories = append(categories, map[string]interface{}{
			"id":          cat.ID,
			"name":        cat.Name,
			"color":       cat.Color,
			"icon":        cat.Icon,
			"is_system":   cat.UserID == nil,
			"parent_id":   cat.ParentID,
			"parent_name": cat.ParentName,
		})
	}

	return response.Success(c, http.StatusOK, categories)
}

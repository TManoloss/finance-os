package handler

import (
	"net/http"
	"strconv"

	"github.com/finance-os/backend/internal/response"
	"github.com/finance-os/backend/internal/service"
	"github.com/labstack/echo/v4"
)

type FeedHandler struct {
	feedService *service.FeedService
}

func NewFeedHandler(feedService *service.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

func (h *FeedHandler) GetFeed(c echo.Context) error {
	userID := c.Get("user_id").(string)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	events, err := h.feedService.GetFeed(c.Request().Context(), userID, page, pageSize)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar feed")
	}

	return response.Success(c, http.StatusOK, events)
}

func (h *FeedHandler) MarkAsRead(c echo.Context) error {
	userID := c.Get("user_id").(string)
	eventID := c.Param("id")

	err := h.feedService.MarkAsRead(c.Request().Context(), userID, eventID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao marcar como lido")
	}

	return response.Success(c, http.StatusOK, map[string]string{"message": "evento marcado como lido"})
}

func (h *FeedHandler) MarkAllAsRead(c echo.Context) error {
	userID := c.Get("user_id").(string)

	err := h.feedService.MarkAllAsRead(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao marcar todos como lidos")
	}

	return response.Success(c, http.StatusOK, map[string]string{"message": "todos os eventos marcados como lidos"})
}

func (h *FeedHandler) GetUnreadCount(c echo.Context) error {
	userID := c.Get("user_id").(string)

	count, err := h.feedService.GetUnreadCount(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "erro ao buscar contagem de não lidos")
	}

	return response.Success(c, http.StatusOK, map[string]int{"unread_count": count})
}

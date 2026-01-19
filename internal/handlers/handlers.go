package handlers

import (
	"golang-test-task/internal/service"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	logger  *slog.Logger
	service service.Service
}

func NewHandlers(logger *slog.Logger, service service.Service) *Handlers {
	return &Handlers{
		logger:  logger,
		service: service,
	}
}

type NumberRequest struct {
	Number *int64 `json:"number"`
}

func (h *Handlers) SaveNumberAndGetSortedNumbers(c echo.Context) error {
	const op = "Handlers.SaveNumberHandler"
	ctx := c.Request().Context()

	var req NumberRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("failed to bind request", slog.Any("op", op), slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.Number == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "number is required",
		})
	}

	if err := h.service.SaveNumber(ctx, *req.Number); err != nil {
		h.logger.Error("failed to save number", slog.Any("op", op), slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot save number"})
	}

	numbers, err := h.service.GetNumbers(ctx)
	if err != nil {
		h.logger.Error("failed to get numbers", slog.Any("op", op), slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot get numbers"})
	}

	h.logger.Debug("numbers saved and fetched successfully", slog.Any("op", op), slog.Int("count", len(numbers)))
	return c.JSON(http.StatusOK, numbers)
}

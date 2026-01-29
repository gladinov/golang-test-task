package handlers

import (
	"golang-test-task/internal/logging"
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
		errMsg := "failed to bind request"
		logging.LoggerError(ctx, h.logger, errMsg, op, err)

		return jsonError(c, http.StatusBadRequest, "invalid request")
	}

	if req.Number == nil {
		return jsonError(c, http.StatusBadRequest, "number is required")
	}

	if err := h.service.SaveNumber(ctx, *req.Number); err != nil {
		errMsg := "failed to save number"
		logging.LoggerError(ctx, h.logger, errMsg, op, err)

		return jsonError(c, http.StatusInternalServerError, "cannot save number")
	}

	numbers, err := h.service.GetNumbers(ctx)
	if err != nil {
		errMsg := "failed to get numbers"
		logging.LoggerError(ctx, h.logger, errMsg, op, err)
		return jsonError(c, http.StatusInternalServerError, "cannot get numbers")
	}

	return c.JSON(http.StatusOK, numbers)
}

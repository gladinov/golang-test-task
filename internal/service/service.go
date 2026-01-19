package service

import (
	"context"
	"fmt"
	"golang-test-task/internal/repository"
	"log/slog"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=Service
type Service interface {
	SaveNumber(ctx context.Context, numb int64) error
	GetNumbers(ctx context.Context) ([]int64, error)
}

type Client struct {
	logger  *slog.Logger
	Storage repository.Storage
}

func New(logger *slog.Logger, storage repository.Storage) *Client {
	return &Client{
		logger:  logger,
		Storage: storage,
	}
}

func (c *Client) SaveNumber(ctx context.Context, numb int64) error {
	const op = "Client.SaveNumber"
	start := time.Now()

	logg := c.logger.With(slog.String("op", op), slog.Int64("number", numb))
	logg.Debug("start saving number")

	if err := c.Storage.SaveNumber(ctx, numb); err != nil {
		logg.Error("failed to save number", slog.Any("error", err))
		return fmt.Errorf("could not save number: %w", err)
	}

	logg.Debug("number saved successfully", slog.Duration("duration", time.Since(start)))
	return nil
}

func (c *Client) GetNumbers(ctx context.Context) ([]int64, error) {
	const op = "Client.GetNumbers"
	start := time.Now()

	logg := c.logger.With(slog.String("op", op))
	logg.Debug("start fetching numbers")

	numbers, err := c.Storage.GetNumbers(ctx)
	if err != nil {
		logg.Error("failed to get numbers", slog.Any("error", err))
		return nil, fmt.Errorf("could not get numbers: %w", err)
	}

	logg.Debug("numbers fetched successfully", slog.Duration("duration", time.Since(start)), slog.Int("count", len(numbers)))
	return numbers, nil
}

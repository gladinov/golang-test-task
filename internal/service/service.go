package service

import (
	"context"
	"golang-test-task/internal/logging"
	"golang-test-task/internal/repository"
	"log/slog"

	"github.com/gladinov/e"
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

func (c *Client) SaveNumber(ctx context.Context, numb int64) (err error) {
	const op = "Client.SaveNumber"

	logg := c.logger.With(slog.Int64("number", numb))
	defer logging.LogOperation_Debug(ctx, logg, op, &err)

	if err := c.Storage.SaveNumber(ctx, numb); err != nil {
		return e.WrapIfErr("could not save number", err)
	}

	return nil
}

func (c *Client) GetNumbers(ctx context.Context) (_ []int64, err error) {
	const op = "Client.GetNumbers"

	defer logging.LogOperation_Debug(ctx, c.logger, op, &err)

	numbers, err := c.Storage.GetNumbers(ctx)
	if err != nil {
		errMsg := "failed to get numbers"
		return nil, e.WrapIfErr(errMsg, err)
	}

	return numbers, nil
}

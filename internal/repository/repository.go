package repository

import (
	"context"
)

type Storage interface {
	SaveNumber(ctx context.Context, numb int64) error
	GetNumbers(ctx context.Context) ([]int64, error)
}

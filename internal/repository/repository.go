package repository

import (
	"context"
)

type Storage interface {
	SaveNumber(ctx context.Context, numb int64) error
	GetNumbers(ctx context.Context) ([]int64, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=DBAdapter
type DBAdapter interface {
	Exec(ctx context.Context, sql string, args ...any) (any, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Close()
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=Rows
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=Row
type Row interface {
	Scan(dest ...any) error
}

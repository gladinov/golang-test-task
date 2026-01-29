package postgreSQL

import (
	"context"
	"golang-test-task/internal/repository"

	"github.com/gladinov/e"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=PoolCreator
type PoolCreator interface {
	NewPool() (repository.DBAdapter, error)
}

type realPoolCreator struct {
	connStr string
}

func NewPoolCreator(connStr string) *realPoolCreator {
	return &realPoolCreator{
		connStr: connStr,
	}
}

func (r *realPoolCreator) NewPool() (repository.DBAdapter, error) {
	pool, err := pgxpool.New(context.Background(), r.connStr)
	if err != nil {
		return nil, e.WrapIfErr("could not create new pgx pool", err)
	}
	return NewPgxAdapter(pool), nil
}

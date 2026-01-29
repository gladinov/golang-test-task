package postgreSQL

import (
	"context"
	"golang-test-task/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxAdapter struct {
	pool *pgxpool.Pool
}

func NewPgxAdapter(pool *pgxpool.Pool) repository.DBAdapter {
	return &pgxAdapter{pool: pool}
}

func (p *pgxAdapter) Exec(ctx context.Context, sql string, args ...any) (any, error) {
	return p.pool.Exec(ctx, sql, args...)
}

func (p *pgxAdapter) Query(ctx context.Context, sql string, args ...any) (repository.Rows, error) {
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRowsAdapter{rows: rows}, nil
}

type pgxRowsAdapter struct {
	rows pgx.Rows
}

func (r *pgxRowsAdapter) Next() bool {
	return r.rows.Next()
}

func (r *pgxRowsAdapter) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *pgxRowsAdapter) Close() {
	r.rows.Close()
}

func (r *pgxRowsAdapter) Err() error {
	return r.rows.Err()
}

// Обертка для QueryRow
type pgxRowAdapter struct {
	row pgx.Row // или pgx.Row, смотря что используешь
}

func (r *pgxRowAdapter) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

func (p *pgxAdapter) QueryRow(ctx context.Context, sql string, args ...any) repository.Row {
	row := p.pool.QueryRow(ctx, sql, args...)
	return &pgxRowAdapter{row: row}
}

func (p *pgxAdapter) Close() {
	p.pool.Close()
}

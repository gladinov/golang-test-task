package postgreSQL

import (
	"context"
	"golang-test-task/internal/logging"
	"golang-test-task/internal/repository"
	"log/slog"

	"github.com/gladinov/e"
)

type Storage struct {
	logger *slog.Logger
	db     repository.DBAdapter
}

func NewStorageWithAdapter(logg *slog.Logger, db repository.DBAdapter) *Storage {
	return &Storage{logger: logg, db: db}
}

func NewPostgresStorageWithCreator(ctx context.Context, logger *slog.Logger, creator PoolCreator) (_ *Storage, err error) {
	const op = "postgreSQL.NewPostgresStorageWithCreator"

	adapter, err := creator.NewPool()
	if err != nil {
		return nil, e.WrapIfErr("could create new postgreSQL storage", err)
	}

	return NewStorageWithAdapter(logger, adapter), nil
}

func MustInitNewStorageWithCreator(ctx context.Context, logger *slog.Logger, creator PoolCreator) *Storage {
	const op = "repository.MustInitNewStorage"
	logg := logger.With(slog.String("op", op))

	serviceStorage, err := NewPostgresStorageWithCreator(ctx, logg, creator)
	if err != nil {
		logg.DebugContext(ctx, "failed to create PostgreSQL storage", "err", err)
		panic(err)
	}

	return serviceStorage
}

func (s *Storage) CloseDB() {
	if s == nil || s.db == nil {
		return
	}
	s.db.Close()
}

func (s *Storage) InitDBForTest(ctx context.Context) (err error) {
	const op = "postgreSQL.InitDBForTest"
	defer func() {
		err = e.WrapIfErr("could not InitDB", err)
	}()
	return s.createNumbersTableForTest(ctx)
}

func (s *Storage) createNumbersTableForTest(ctx context.Context) error {
	const op = "postgreSQL.createNumbersTableForTest"
	_, err := s.db.Exec(ctx, queryCreateNumberTable)
	if err != nil {
		return e.WrapIfErr("could not create nubmers table", err)
	}
	return nil
}

func (s *Storage) SaveNumber(
	ctx context.Context,
	value int64,
) (err error) {
	const op = "postgreSql.SaveNumber"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	defer func() {
		err = e.WrapIfErr("can't save number", err)
	}()
	_, err = s.db.Exec(ctx, `
		INSERT INTO numbers (number)
		VALUES ($1)
	`, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetNumbers(ctx context.Context) (_ []int64, err error) {
	const op = "postgreSql.GetNumbers"

	defer logging.LogOperation_Debug(ctx, s.logger, op, &err)()

	rows, err := s.db.Query(ctx, `
		SELECT number
		FROM numbers
		ORDER BY number ASC
	`)
	if err != nil {
		return nil, e.WrapIfErr("failed to query numbers", err)
	}
	defer rows.Close()

	var res []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, e.WrapIfErr("failed to scan number", err)
		}
		res = append(res, v)
	}

	if err := rows.Err(); err != nil {
		return nil, e.WrapIfErr("rows error", err)
	}

	return res, nil
}

func (s *Storage) TruncateNumbers(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `DROP TABLE IF EXISTS numbers`)
	return err
}

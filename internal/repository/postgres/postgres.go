package postgreSQL

import (
	"context"
	"fmt"
	"golang-test-task/internal/repository"
	"log/slog"
	"time"

	"github.com/gladinov/e"
)

type Storage struct {
	logger *slog.Logger
	db     repository.DBAdapter
}

func NewStorageWithAdapter(logg *slog.Logger, db repository.DBAdapter) *Storage {
	return &Storage{logger: logg, db: db}
}

func MustInitNewStorageWithCreator(ctx context.Context, logg *slog.Logger, creator PoolCreator) *Storage {
	const op = "repository.MustInitNewStorage"
	logg.Debug(fmt.Sprintf("start %s", op))

	serviceStorage, err := NewPostgresStorageWithCreator(logg, creator)
	if err != nil {
		logg.Debug("failed to create PostgreSQL storage", "err", err)
		panic(err)
	}
	err = serviceStorage.InitDB(ctx)
	if err != nil {
		logg.Debug("failed to init PostgreSQL database", "err", err)
		panic(err)
	}
	logg.Info("PostgreSQL storage initialized successfully")
	return serviceStorage
}

func NewPostgresStorageWithCreator(logger *slog.Logger, creator PoolCreator) (_ *Storage, err error) {
	const op = "postgreSQL.NewPostgresStorageWithCreator"
	defer func() {
		err = e.WrapIfErr("could create new postgreSQL storage", err)
	}()
	adapter, err := creator.NewPool()
	if err != nil {
		return nil, err
	}

	return NewStorageWithAdapter(logger, adapter), nil
}

func (s *Storage) CloseDB() {
	if s == nil || s.db == nil {
		return
	}
	s.db.Close()
}

func (s *Storage) InitDB(ctx context.Context) (err error) {
	const op = "postgreSQL.InitDB"
	defer func() {
		err = e.WrapIfErr("could not InitDB", err)
	}()
	return s.createNumbersTable(ctx)
}

func (s *Storage) createNumbersTable(ctx context.Context) error {
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

	start := time.Now()
	logg := s.logger.With(slog.String("op", op))
	logg.Debug("start")
	defer func() {
		logg.Debug("finished",
			slog.Duration("duration", time.Since(start)),
			slog.Any("error", err),
		)
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

	start := time.Now()
	logg := s.logger.With(slog.String("op", op))
	logg.Debug("start")
	defer func() {
		logg.Debug("finished",
			slog.Duration("duration", time.Since(start)),
		)

		err = e.WrapIfErr("can't get number", err)
	}()

	rows, err := s.db.Query(ctx, `
		SELECT number
		FROM numbers
		ORDER BY number ASC
	`)
	if err != nil {
		return nil, e.WrapIfErr("query numbers", err)
	}
	defer rows.Close()

	var res []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, e.WrapIfErr("scan number", err)
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

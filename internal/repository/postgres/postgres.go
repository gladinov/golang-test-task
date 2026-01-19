package postgreSQL

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	config "golang-test-task/internal/configs"

	"github.com/gladinov/e"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	logger *slog.Logger
	db     *pgxpool.Pool
}

func NewStorage(logger *slog.Logger, postgresConfig config.Config) (_ *Storage, err error) {
	const op = "postgreSQL.NewStorage"

	start := time.Now()
	logg := logger.With(
		slog.String("op", op))
	logg.Debug(fmt.Sprintf("start %s", op))
	defer func() {
		logg.Debug("fineshed",
			slog.Duration("duration", time.Since(start)),
			slog.Any("error", err),
		)

		err = e.WrapIfErr("could create new postgreSQL storage", err)
	}()

	postgresHost, err := postgresConfig.PostgresHost.GetStringHost()
	if err != nil {
		return nil, err
	}
	db, err := pgxpool.New(context.Background(), postgresHost)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db, logger: logger}, nil
}

func (s *Storage) CloseDB() {
	s.db.Close()
}

func (s *Storage) InitDB(ctx context.Context) (err error) {
	const op = "postgreSQL.InitDB"

	start := time.Now()
	logg := s.logger.With(
		slog.String("op", op))
	logg.Debug(fmt.Sprintf("start %s", op))
	defer func() {
		logg.Debug("fineshed",
			slog.Duration("duration", time.Since(start)),
			slog.Any("error", err),
		)
		err = e.WrapIfErr("could not InitDB", err)
	}()
	err = s.createNumbersTable(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) createNumbersTable(ctx context.Context) error {
	_, err := s.db.Exec(ctx, queryCreateNumberTable)
	if err != nil {
		return e.WrapIfErr("could not create bond reports table", err)
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
		return fmt.Errorf("insert number: %w", err)
	}

	return nil
}

func (s *Storage) GetNumbers(ctx context.Context) ([]int64, error) {
	const op = "postgreSql.GetNumbers"

	start := time.Now()
	logg := s.logger.With(slog.String("op", op))
	logg.Debug("start")
	defer func() {
		logg.Debug("finished",
			slog.Duration("duration", time.Since(start)),
		)
	}()

	rows, err := s.db.Query(ctx, `
		SELECT number
		FROM numbers
		ORDER BY number ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query numbers: %w", err)
	}
	defer rows.Close()

	var res []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan number: %w", err)
		}
		res = append(res, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return res, nil
}

func MustInitNewStorage(ctx context.Context, config config.Config, logg *slog.Logger) *Storage {
	const op = "repository.MustInitNewStorage"
	logg.Debug(fmt.Sprintf("start %s", op))

	serviceStorage, err := NewStorage(logg, config)
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

func (s *Storage) TruncateNumbers(ctx context.Context) error {
	_, err := s.db.Exec(ctx, "DELETE FROM numbers")
	return err
}

//go:build integration

package postgreSQL

import (
	"context"
	config "golang-test-task/internal/configs"
	"golang-test-task/internal/testutils/rootpathfinder"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	rootpathfinder.MustInitialize()
	rootPath := rootpathfinder.MustGetRoot()
	envsPath := "deployments/envs/test.env"

	abspath := filepath.Join(rootPath, envsPath)

	if err := godotenv.Load(abspath); err != nil {
		panic("failed to load test env file: " + err.Error())
	}
	os.Setenv("ROOT_PATH", rootPath)

	code := m.Run()

	os.Exit(code)
}

func NewTestStorage(t *testing.T) *Storage {
	t.Helper()
	ctx := context.Background()

	conf := config.MustInitConfig()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	postgresHost, err := conf.PostgresHost.GetStringHost()
	require.NoError(t, err)

	creator := NewPoolCreator(postgresHost)

	storage, err := NewPostgresStorageWithCreator(ctx, logger, creator)
	require.NoError(t, err)

	err = storage.InitDB(ctx)
	require.NoError(t, err)

	err = storage.TruncateNumbers(context.Background())
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = storage.TruncateNumbers(context.Background())
		storage.CloseDB()
	})

	return storage
}

func TestNewStorage_Integration_Success(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	conf := config.MustInitConfig()
	postgresHost, err := conf.PostgresHost.GetStringHost()
	require.NoError(t, err)

	creator := NewPoolCreator(postgresHost)

	storage, err := NewPostgresStorageWithCreator(ctx, logger, creator)
	require.NoError(t, err)
	require.NoError(t, err)
	require.NotNil(t, storage)
	require.NotNil(t, storage.db)

	t.Cleanup(func() {
		storage.CloseDB()
	})

	var one int
	err = storage.db.QueryRow(context.Background(), "SELECT 1").Scan(&one)
	require.NoError(t, err)
	require.Equal(t, 1, one)
}

func TestInitDB_Integration(t *testing.T) {
	ctx := context.Background()
	storage := NewTestStorage(t)
	err := storage.InitDB(ctx)
	require.NoError(t, err)

	var exists bool
	err = storage.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema='public' AND table_name='numbers'
		)
	`).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestSaveNumber_Integration(t *testing.T) {
	ctx := context.Background()
	storage := NewTestStorage(t)
	err := storage.InitDB(ctx)
	require.NoError(t, err)

	var number int64 = 1
	err = storage.SaveNumber(ctx, number)
	require.NoError(t, err)

	var res int64
	err = storage.db.QueryRow(ctx, `SELECT number FROM numbers LIMIT 1`).Scan(&res)
	require.NoError(t, err)
	require.Equal(t, number, res)
}

func TestSaveSeveralNumbers_Integration(t *testing.T) {
	ctx := context.Background()
	storage := NewTestStorage(t)
	err := storage.InitDB(ctx)
	require.NoError(t, err)
	for i := int64(1); i <= 3; i++ {
		require.NoError(t, storage.SaveNumber(ctx, i))
	}

	rows, err := storage.db.Query(ctx, `SELECT number FROM numbers ORDER BY number`)
	require.NoError(t, err)
	defer rows.Close()

	var got []int64
	for rows.Next() {
		var n int64
		require.NoError(t, rows.Scan(&n))
		got = append(got, n)
	}
	require.Equal(t, []int64{1, 2, 3}, got)
}

func TestGetNumbers_Integration(t *testing.T) {
	ctx := context.Background()
	storage := NewTestStorage(t)
	err := storage.InitDB(ctx)
	require.NoError(t, err)

	var number int64 = 1
	err = storage.SaveNumber(ctx, number)
	require.NoError(t, err)

	res, err := storage.GetNumbers(ctx)
	require.NoError(t, err)
	require.Equal(t, []int64{number}, res)
}

func TestGetSeveralNumbers_Integration(t *testing.T) {
	ctx := context.Background()
	storage := NewTestStorage(t)
	err := storage.InitDB(ctx)
	require.NoError(t, err)

	numbers := []int64{1, 3, 2}
	for _, number := range numbers {
		require.NoError(t, storage.SaveNumber(ctx, number))
	}

	want := []int64{1, 2, 3}

	res, err := storage.GetNumbers(ctx)
	require.NoError(t, err)
	require.Equal(t, want, res)
}

func TestGetNumbers_Empty_Integration(t *testing.T) {
	ctx := context.Background()
	storage := NewTestStorage(t)
	err := storage.InitDB(ctx)
	require.NoError(t, err)

	res, err := storage.GetNumbers(ctx)
	require.NoError(t, err)
	require.Empty(t, res)
}

func TestMustInitStorage_Integration(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	conf := config.MustInitConfig()
	postgresHost, _ := conf.PostgresHost.GetStringHost()
	creator := NewPoolCreator(postgresHost)
	storage := MustInitNewStorageWithCreator(ctx, logger, creator)

	require.NotNil(t, storage)
	require.NotNil(t, storage.db)

	t.Cleanup(func() {
		storage.CloseDB()
	})

	var one int
	err := storage.db.QueryRow(context.Background(), "SELECT 1").Scan(&one)
	require.NoError(t, err)
	require.Equal(t, 1, one)
}

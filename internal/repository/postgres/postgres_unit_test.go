//go:build unit

package postgreSQL

import (
	"context"
	"errors"
	"fmt"
	"golang-test-task/internal/repository/mocks"
	"io"
	"log/slog"
	"strings"
	"testing"

	poolCreatorMocks "golang-test-task/internal/repository/postgres/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInitDB_Unit(t *testing.T) {
	logg := slog.New(slog.NewTextHandler(io.Discard, nil))

	cases := []struct {
		name      string
		wantErr   bool
		setupMock func(*mocks.DBAdapter)
		err       error
	}{
		{
			name:    "success",
			wantErr: false,
			setupMock: func(mockDBAdapter *mocks.DBAdapter) {
				mockDBAdapter.On("Exec", mock.Anything,
					mock.MatchedBy(func(q string) bool {
						q = strings.ToLower(q)
						return strings.Contains(q, "create table if not exists numbers") &&
							strings.Contains(q, "created_at") &&
							strings.Contains(q, "timestamptz") &&
							strings.Contains(q, "number") &&
							strings.Contains(q, "bigint not null")
					}),
					mock.Anything,
				).Return(nil, nil).Once()
			},
			err: nil,
		},
		{
			name:    "db error",
			wantErr: true,
			setupMock: func(mockDBAdapter *mocks.DBAdapter) {
				mockDBAdapter.
					On("Exec",
						mock.Anything,
						mock.MatchedBy(func(q string) bool {
							q = strings.ToLower(q)
							return strings.Contains(q, "create table if not exists numbers")
						}),
						mock.Anything,
					).
					Return(nil, fmt.Errorf("db down")).
					Once()
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockDBAdapter := mocks.NewDBAdapter(t)
			storage := NewStorageWithAdapter(logg, mockDBAdapter)
			tc.setupMock(mockDBAdapter)
			err := storage.InitDB(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "could not InitDB")
				require.Contains(t, err.Error(), "could not create nubmers table")
			} else {
				require.NoError(t, err)
			}
			mockDBAdapter.AssertExpectations(t)
		})
	}
}

func TestNewPostgresStorageWithCreator_Unit(t *testing.T) {
	ctx := context.Background()
	logg := slog.New(slog.NewTextHandler(io.Discard, nil))

	cases := []struct {
		name string

		setupPoolMock func(*poolCreatorMocks.PoolCreator)
		wantErr       bool
		err           error
	}{
		{
			name: "success",
			setupPoolMock: func(mockPoolCreator *poolCreatorMocks.PoolCreator) {
				mockPoolCreator.On("NewPool").Return(nil, nil).Once()
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "invalid host",
			setupPoolMock: func(mockPoolCreator *poolCreatorMocks.PoolCreator) {
				mockPoolCreator.On("NewPool").Return(nil, errors.New("could not create new pgx pool")).Once()
			},
			wantErr: true,
			err:     errors.New("could not create new pgx pool"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockPoolCreator := poolCreatorMocks.NewPoolCreator(t)

			if tc.setupPoolMock != nil {
				tc.setupPoolMock(mockPoolCreator)
			}
			storage, err := NewPostgresStorageWithCreator(ctx, logg, mockPoolCreator)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, storage)
				require.Contains(t, err.Error(), tc.err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, storage)
			}
			mockPoolCreator.AssertExpectations(t)
		})
	}
}

func TestSaveNumber_Unit(t *testing.T) {
	logg := slog.New(slog.NewTextHandler(io.Discard, nil))

	ctx := context.Background()
	value := int64(42)

	cases := []struct {
		name      string
		value     int64
		setupMock func(*mocks.DBAdapter)
		wantErr   bool
	}{
		{
			name:  "success",
			value: value,
			setupMock: func(db *mocks.DBAdapter) {
				db.On(
					"Exec",
					mock.Anything, // context
					mock.MatchedBy(func(q string) bool {
						q = strings.ToLower(q)
						return strings.Contains(q, "insert into numbers") &&
							strings.Contains(q, "values ($1)")
					}),
					value,
				).Return(nil, nil).Once()
			},
			wantErr: false,
		},
		{
			name:  "db error",
			value: value,
			setupMock: func(db *mocks.DBAdapter) {
				db.On(
					"Exec",
					mock.Anything,
					mock.Anything,
					value,
				).Return(nil, errors.New("db is down")).Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := mocks.NewDBAdapter(t)
			tc.setupMock(db)

			storage := NewStorageWithAdapter(logg, db)

			err := storage.SaveNumber(ctx, tc.value)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "can't save number")
			} else {
				require.NoError(t, err)
			}

			db.AssertExpectations(t)
		})
	}
}

func TestGetNumbers_Unit(t *testing.T) {
	logg := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	cases := []struct {
		name            string
		setupMock       func(db *mocks.DBAdapter, rows *mocks.Rows)
		want            []int64
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "success",
			setupMock: func(db *mocks.DBAdapter, rows *mocks.Rows) {
				db.On(
					"Query",
					mock.Anything,
					mock.MatchedBy(func(q string) bool {
						q = strings.ToLower(q)
						return strings.Contains(q, "select number") &&
							strings.Contains(q, "from numbers") &&
							strings.Contains(q, "order by number")
					}),
				).Return(rows, nil).Once()

				rows.On("Next").Return(true).Once()
				rows.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
					ptr := args.Get(0).(*int64)
					*ptr = 1
				}).Return(nil).Once()

				rows.On("Next").Return(true).Once()
				rows.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
					ptr := args.Get(0).(*int64)
					*ptr = 2
				}).Return(nil).Once()

				rows.On("Next").Return(false).Once()
				rows.On("Err").Return(nil).Once()
				rows.On("Close").Return().Once()
			},
			want:    []int64{1, 2},
			wantErr: false,
		},
		{
			name: "query error",
			setupMock: func(db *mocks.DBAdapter, _ *mocks.Rows) {
				db.On(
					"Query",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("query failed")).Once()
			},
			wantErr:         true,
			wantErrContains: "failed to query numbers",
		},
		{
			name: "scan error",
			setupMock: func(db *mocks.DBAdapter, rows *mocks.Rows) {
				db.On("Query", mock.Anything, mock.Anything).
					Return(rows, nil).Once()

				rows.On("Next").Return(true).Once()
				rows.On("Scan", mock.Anything).
					Return(errors.New("scan failed")).Once()
				rows.On("Close").Return().Once()
			},
			wantErr:         true,
			wantErrContains: "failed to scan number",
		},
		{
			name: "rows.Err returns error",
			setupMock: func(db *mocks.DBAdapter, rows *mocks.Rows) {
				numbers := []int64{1, 2}
				i := 0

				db.On("Query", mock.Anything, mock.Anything).Return(rows, nil).Once()

				rows.On("Next").Return(func() bool {
					if i < len(numbers) {
						return true
					}
					return false
				}).Times(len(numbers) + 1) // +1 для последнего Next() == false

				rows.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
					ptr := args.Get(0).(*int64)
					*ptr = numbers[i]
					i++
				}).Return(nil).Times(len(numbers))

				rows.On("Err").Return(errors.New("rows error")).Once()
				rows.On("Close").Return().Once()
			},
			wantErr:         true,
			wantErrContains: "rows error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := mocks.NewDBAdapter(t)
			rows := mocks.NewRows(t)

			tc.setupMock(db, rows)

			storage := NewStorageWithAdapter(logg, db)

			res, err := storage.GetNumbers(ctx)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, res)
			}

			db.AssertExpectations(t)
			rows.AssertExpectations(t)
		})
	}
}

func TestTruncateNumbers_Unit(t *testing.T) {
	logg := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	cases := []struct {
		name      string
		setupMock func(db *mocks.DBAdapter)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(db *mocks.DBAdapter) {
				db.On("Exec", mock.Anything, mock.MatchedBy(func(q string) bool {
					return strings.Contains(strings.ToLower(q), "drop table if exists numbers")
				})).Return(nil, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "exec error",
			setupMock: func(db *mocks.DBAdapter) {
				db.On("Exec", mock.Anything, mock.Anything).
					Return(nil, errors.New("exec failed")).Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := mocks.NewDBAdapter(t)
			if tc.setupMock != nil {
				tc.setupMock(db)
			}

			storage := NewStorageWithAdapter(logg, db)
			err := storage.TruncateNumbers(ctx)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "exec failed")
			} else {
				require.NoError(t, err)
			}

			db.AssertExpectations(t)
		})
	}
}

func TestMustInitNewStorageWithCreator(t *testing.T) {
	ctx := context.Background()
	logg := slog.New(slog.NewTextHandler(io.Discard, nil))

	cases := []struct {
		name          string
		setupMock     func(creator *poolCreatorMocks.PoolCreator, db *mocks.DBAdapter)
		expectPanic   bool
		panicContains string
	}{
		{
			name: "success",
			setupMock: func(creator *poolCreatorMocks.PoolCreator, db *mocks.DBAdapter) {
				db.On("Exec", mock.Anything, mock.Anything).Return(nil, nil).Once()
				creator.On("NewPool").Return(db, nil).Once()
			},
			expectPanic: false,
		},
		{
			name: "creator.NewPool returns error -> panic",
			setupMock: func(creator *poolCreatorMocks.PoolCreator, db *mocks.DBAdapter) {
				creator.On("NewPool").Return(nil, errors.New("cannot create pool")).Once()
			},
			expectPanic:   true,
			panicContains: "cannot create pool",
		},
		{
			name: "InitDB returns error -> panic",
			setupMock: func(creator *poolCreatorMocks.PoolCreator, db *mocks.DBAdapter) {
				db.On("Exec", mock.Anything, mock.Anything).Return(nil, errors.New("init db failed")).Once()
				creator.On("NewPool").Return(db, nil).Once()
			},
			expectPanic:   true,
			panicContains: "init db failed",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := mocks.NewDBAdapter(t)
			creator := poolCreatorMocks.NewPoolCreator(t)

			tc.setupMock(creator, db)

			if tc.expectPanic {
				func() {
					defer func() {
						if r := recover(); r != nil {
							if !strings.Contains(fmt.Sprint(r), tc.panicContains) {
								t.Fatalf("panic value does not contain expected substring: %q, got: %v", tc.panicContains, r)
							}
						} else {
							t.Fatal("expected panic, but function did not panic")
						}
					}()
					MustInitNewStorageWithCreator(ctx, logg, creator)
				}()
			} else {
				storage := MustInitNewStorageWithCreator(ctx, logg, creator)
				require.NotNil(t, storage)
			}

			db.AssertExpectations(t)
			creator.AssertExpectations(t)
		})
	}
}

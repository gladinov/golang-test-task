//go:build unit

package service

import (
	"context"
	"errors"
	"golang-test-task/internal/service/mocks"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_SaveNumber(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockStorage := mocks.NewService(t)
	client := New(logger, mockStorage)

	mockStorage.On("SaveNumber", ctx, int64(42)).Return(nil)

	err := client.SaveNumber(ctx, 42)
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)

	mockStorage.On("SaveNumber", ctx, int64(7)).Return(errors.New("db error"))

	err = client.SaveNumber(ctx, 7)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not save number")
	mockStorage.AssertExpectations(t)
}

func TestClient_GetNumbers(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // логгер в тесте

	mockStorage := mocks.NewService(t)
	client := New(logger, mockStorage)

	// Определяем набор тестов
	testCases := []struct {
		name       string
		mockReturn []int64
		mockError  error
		wantNums   []int64
		wantErr    bool
	}{
		{
			name:       "success",
			mockReturn: []int64{1, 2, 3},
			mockError:  nil,
			wantNums:   []int64{1, 2, 3},
			wantErr:    false,
		},
		{
			name:       "db error",
			mockReturn: nil,
			mockError:  errors.New("db error"),
			wantNums:   nil,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage.On("GetNumbers", ctx).Return(tc.mockReturn, tc.mockError)

			numbers, err := client.GetNumbers(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, numbers)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantNums, numbers)
			}

			mockStorage.AssertExpectations(t)
			mockStorage.ExpectedCalls = nil
			mockStorage.Calls = nil
		})
	}
}

package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"golang-test-task/internal/service/mocks"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlers_SaveNumberAndGetSortedNumbers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockService := mocks.NewService(t)
	h := NewHandlers(logger, mockService)

	e := echo.New()

	testCases := []struct {
		name           string
		inputBody      map[string]any
		mockSaveErr    error
		mockGetNums    []int64
		mockGetErr     error
		expectedStatus int
		expectedBody   any
	}{
		{
			name:           "success",
			inputBody:      map[string]any{"number": 42},
			mockSaveErr:    nil,
			mockGetNums:    []int64{1, 2, 42},
			mockGetErr:     nil,
			expectedStatus: http.StatusOK,
			expectedBody:   []int64{1, 2, 42},
		},
		{
			name:           "invalid request",
			inputBody:      map[string]any{"num": 42},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "number is required"},
		},
		{
			name:           "invalid request with string",
			inputBody:      map[string]any{"number": "42"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "invalid request"},
		},
		{
			name:           "save number error",
			inputBody:      map[string]any{"number": 42},
			mockSaveErr:    errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "cannot save number"},
		},
		{
			name:           "get numbers error",
			inputBody:      map[string]any{"number": 42},
			mockSaveErr:    nil,
			mockGetNums:    nil,
			mockGetErr:     errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "cannot get numbers"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			bodyBytes, _ := json.Marshal(tc.inputBody)
			req := httptest.NewRequest(http.MethodPost, "/number", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)

			if tc.name != "invalid request" && tc.name != "invalid request with string" {
				mockService.On("SaveNumber", mock.Anything, mock.Anything).Return(tc.mockSaveErr)
				if tc.mockSaveErr == nil {
					mockService.On("GetNumbers", mock.Anything).Return(tc.mockGetNums, tc.mockGetErr)
				}
			}

			err := h.SaveNumberAndGetSortedNumbers(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if rec.Code == http.StatusOK {
				var respBody []int64
				err := json.Unmarshal(rec.Body.Bytes(), &respBody)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, respBody)
			} else {
				var respBody map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &respBody)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, respBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

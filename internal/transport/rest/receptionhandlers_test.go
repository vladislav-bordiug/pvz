package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvz/internal/contextkeys"
	"pvz/internal/models"
)

func TestCloseLastReceptionHandler(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name             string
		urlVars          map[string]string
		role             string
		serviceSetup     func(m *MockService)
		expectedStatus   int
		expectedResponse func(body []byte)
	}{
		{
			name:    "invalid pvz id",
			urlVars: map[string]string{"pvzId": "invalid-uuid"},
			role:    "employee",

			serviceSetup:   func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: func(body []byte) {
				var errResp models.ErrorResponse
				err := json.Unmarshal(body, &errResp)
				assert.NoError(t, err)
				assert.Equal(t, "Неверный идентификатор ПВЗ", errResp.Message)
			},
		},
		{
			name:    "service returns error",
			urlVars: map[string]string{"pvzId": validUUID.String()},
			role:    "employee",
			serviceSetup: func(m *MockService) {
				m.
					On("CloseLastReception", mock.Anything, "employee", validUUID).
					Return(nil, http.StatusBadRequest, errors.New("ошибка закрытия приёмки"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: func(body []byte) {
				var errResp models.ErrorResponse
				err := json.Unmarshal(body, &errResp)
				assert.NoError(t, err)
				assert.Equal(t, "ошибка закрытия приёмки", errResp.Message)
			},
		},
		{
			name:    "success",
			urlVars: map[string]string{"pvzId": validUUID.String()},
			role:    "employee",
			serviceSetup: func(m *MockService) {
				expectedRec := &models.Reception{
					ID:       validUUID,
					Status:   "close",
					DateTime: time.Now(),
					PVZId:    validUUID,
				}
				m.
					On("CloseLastReception", mock.Anything, "employee", validUUID).
					Return(expectedRec, http.StatusOK, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: func(body []byte) {
				var rec models.Reception
				err := json.Unmarshal(body, &rec)
				assert.NoError(t, err)
				assert.Equal(t, "close", rec.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodPost, "/close-last-reception", nil)

			req = mux.SetURLVars(req, tt.urlVars)

			req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, tt.role))
			rr := httptest.NewRecorder()

			mockSvc := new(MockService)
			tt.serviceSetup(mockSvc)

			handler := NewHandler(mockSvc)
			handler.CloseLastReceptionHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedResponse(rr.Body.Bytes())
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestCreateReceptionHandler(t *testing.T) {

	validUUID := uuid.New()

	tests := []struct {
		name             string
		requestBody      string
		role             string
		serviceSetup     func(m *MockService)
		expectedStatus   int
		expectedResponse func(body []byte)
	}{
		{
			name:        "invalid json",
			requestBody: "invalid json",
			role:        "employee",
			serviceSetup: func(m *MockService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: func(body []byte) {
				var errResp models.ErrorResponse
				err := json.Unmarshal(body, &errResp)
				assert.NoError(t, err)
				assert.Equal(t, "Неверный запрос", errResp.Message)
			},
		},
		{
			name:        "invalid pvz id in request",
			requestBody: `{"PVZId": "invalid-uuid"}`,
			role:        "employee",
			serviceSetup: func(m *MockService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: func(body []byte) {
				var errResp models.ErrorResponse
				err := json.Unmarshal(body, &errResp)
				assert.NoError(t, err)
				assert.Equal(t, "Неверный идентификатор ПВЗ", errResp.Message)
			},
		},
		{
			name: "service returns error",
			requestBody: func() string {
				req := models.CreateReceptionRequest{PVZId: validUUID.String()}
				b, _ := json.Marshal(req)
				return string(b)
			}(),
			role: "employee",
			serviceSetup: func(m *MockService) {
				m.
					On("CreateReception", mock.Anything, "employee", validUUID).
					Return(nil, http.StatusBadRequest, errors.New("ошибка создания приёмки"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: func(body []byte) {
				var errResp models.ErrorResponse
				err := json.Unmarshal(body, &errResp)
				assert.NoError(t, err)
				assert.Equal(t, "ошибка создания приёмки", errResp.Message)
			},
		},
		{
			name: "success",
			requestBody: func() string {
				req := models.CreateReceptionRequest{PVZId: validUUID.String()}
				b, _ := json.Marshal(req)
				return string(b)
			}(),
			role: "employee",
			serviceSetup: func(m *MockService) {
				expectedRec := &models.Reception{
					ID:       validUUID,
					Status:   "in_progress",
					DateTime: time.Now(),
					PVZId:    validUUID,
				}
				m.
					On("CreateReception", mock.Anything, "employee", validUUID).
					Return(expectedRec, http.StatusOK, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: func(body []byte) {
				var rec models.Reception
				err := json.Unmarshal(body, &rec)
				assert.NoError(t, err)
				assert.Equal(t, "in_progress", rec.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/create-reception", bytes.NewBufferString(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, tt.role))
			rr := httptest.NewRecorder()

			mockSvc := new(MockService)
			tt.serviceSetup(mockSvc)

			handler := NewHandler(mockSvc)
			handler.CreateReceptionHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedResponse(rr.Body.Bytes())
			mockSvc.AssertExpectations(t)
		})
	}
}

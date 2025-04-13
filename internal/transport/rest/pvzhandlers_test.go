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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvz/internal/contextkeys"
	"pvz/internal/models"
)

func TestCreatePVZHandler(t *testing.T) {
	handlerTests := []struct {
		name           string
		requestBody    string
		role           string
		mockSetup      func(m *MockService)
		expectedStatus int
		expectedBody   func(body []byte)
	}{
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			role:           "moderator",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(body []byte) {
				var errResp models.ErrorResponse
				assert.NoError(t, json.Unmarshal(body, &errResp))
				assert.Equal(t, "Неверный запрос", errResp.Message)
			},
		},
		{
			name: "service returns error",
			requestBody: func() string {
				req := models.PVZ{City: "NotValidCity"}
				b, _ := json.Marshal(req)
				return string(b)
			}(),
			role: "moderator",
			mockSetup: func(m *MockService) {
				errorMessage := "ошибка ПВЗ можно создать только в Москве, Санкт-Петербурге или Казани"
				m.On("CreatePVZ", mock.Anything, mock.MatchedBy(func(p *models.PVZ) bool {
					return p.City == "NotValidCity"
				}), "moderator").Return(http.StatusBadRequest, errors.New(errorMessage))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(body []byte) {
				var errResp models.ErrorResponse
				assert.NoError(t, json.Unmarshal(body, &errResp))
				assert.Equal(t, "ошибка ПВЗ можно создать только в Москве, Санкт-Петербурге или Казани", errResp.Message)
			},
		},
		{
			name: "success",
			requestBody: func() string {
				req := models.PVZ{City: "Москва"}
				b, _ := json.Marshal(req)
				return string(b)
			}(),
			role: "moderator",
			mockSetup: func(m *MockService) {
				m.On("CreatePVZ", mock.Anything, mock.MatchedBy(func(p *models.PVZ) bool {
					return p.City == "Москва"
				}), "moderator").Return(http.StatusOK, nil).Run(func(args mock.Arguments) {
					p := args.Get(1).(*models.PVZ)
					p.RegistrationDate = time.Now()
					p.ID = uuid.New()
				})
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(body []byte) {
				var pvzResp models.PVZ
				assert.NoError(t, json.Unmarshal(body, &pvzResp))
				assert.Equal(t, "Москва", pvzResp.City)
				assert.NotEqual(t, uuid.Nil, pvzResp.ID)
				assert.False(t, pvzResp.RegistrationDate.IsZero())
			},
		},
	}

	for _, tt := range handlerTests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/create-pvz", bytes.NewBufferString(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, tt.role))
			rr := httptest.NewRecorder()

			mockSvc := new(MockService)
			tt.mockSetup(mockSvc)

			handler := NewHandler(mockSvc)
			handler.CreatePVZHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedBody(rr.Body.Bytes())
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestListPVZHandler(t *testing.T) {
	handlerTests := []struct {
		name           string
		queryString    string
		mockSetup      func(m *MockService)
		expectedStatus int
		expectedBody   func(body []byte)
	}{
		{
			name:        "service returns error",
			queryString: "/list-pvz?startDate=2023-01-01T00:00:00Z&endDate=2023-01-02T00:00:00Z&page=1&limit=10",
			mockSetup: func(m *MockService) {
				errorMessage := "ошибка выборки ПВЗ"
				m.On("ListPVZ", mock.Anything, "2023-01-01T00:00:00Z", "2023-01-02T00:00:00Z", "1", "10").
					Return(([]*models.PVZResponse)(nil), http.StatusBadRequest, errors.New(errorMessage))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(body []byte) {
				var errResp models.ErrorResponse
				assert.NoError(t, json.Unmarshal(body, &errResp))
				assert.Equal(t, "ошибка выборки ПВЗ", errResp.Message)
			},
		},
		{
			name:        "success with default dates",
			queryString: "/list-pvz?page=2&limit=5",
			mockSetup: func(m *MockService) {
				samplePVZ := models.PVZ{
					ID:               uuid.New(),
					City:             "Москва",
					RegistrationDate: time.Now(),
				}
				sampleResponse := &models.PVZResponse{
					PVZ:        &samplePVZ,
					Receptions: []*models.ReceptionInfo{},
				}
				m.On("ListPVZ", mock.Anything, "", "", "2", "5").
					Return([]*models.PVZResponse{sampleResponse}, http.StatusOK, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(body []byte) {
				var results []*models.PVZResponse
				assert.NoError(t, json.Unmarshal(body, &results))
				assert.Len(t, results, 1)
				assert.Equal(t, "Москва", results[0].PVZ.City)
			},
		},
	}

	for _, tt := range handlerTests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.queryString, nil)
			rr := httptest.NewRecorder()

			mockSvc := new(MockService)
			tt.mockSetup(mockSvc)

			handler := NewHandler(mockSvc)
			handler.ListPVZHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedBody(rr.Body.Bytes())
			mockSvc.AssertExpectations(t)
		})
	}
}

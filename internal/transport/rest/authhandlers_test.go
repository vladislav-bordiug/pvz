package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	pb "pvz/internal/pb/pvz_v1"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvz/internal/models"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) DummyLogin(req *models.DummyLoginRequest) (string, int, error) {
	args := m.Called(req)
	return args.String(0), args.Int(1), args.Error(2)
}

func (m *MockService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, int, error) {
	args := m.Called(ctx, req)
	var user *models.User
	if u := args.Get(0); u != nil {
		user = u.(*models.User)
	}
	return user, args.Int(1), args.Error(2)
}

func (m *MockService) Login(ctx context.Context, req *models.LoginRequest) (string, int, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Int(1), args.Error(2)
}

func (m *MockService) CreatePVZ(ctx context.Context, pvz *models.PVZ, role string) (int, error) {
	args := m.Called(ctx, pvz, role)
	return args.Int(0), args.Error(1)
}

func (m *MockService) ListPVZ(ctx context.Context, startDateStr, endDateStr, pageStr, limitStr string) ([]*models.PVZResponse, int, error) {
	args := m.Called(ctx, startDateStr, endDateStr, pageStr, limitStr)
	var res []*models.PVZResponse
	if args.Get(0) != nil {
		res = args.Get(0).([]*models.PVZResponse)
	}
	return res, args.Int(1), args.Error(2)
}

func (m *MockService) CloseLastReception(ctx context.Context, role string, pvzId uuid.UUID) (*models.Reception, int, error) {
	args := m.Called(ctx, role, pvzId)
	var rec *models.Reception
	if r := args.Get(0); r != nil {
		rec = r.(*models.Reception)
	}
	return rec, args.Int(1), args.Error(2)
}

func (m *MockService) DeleteLastProduct(ctx context.Context, role string, pvzId uuid.UUID) (int, error) {
	args := m.Called(ctx, role, pvzId)
	return args.Int(0), args.Error(1)
}

func (m *MockService) CreateReception(ctx context.Context, role string, pvzId uuid.UUID) (*models.Reception, int, error) {
	args := m.Called(ctx, role, pvzId)
	var rec *models.Reception
	if r := args.Get(0); r != nil {
		rec = r.(*models.Reception)
	}
	return rec, args.Int(1), args.Error(2)
}

func (m *MockService) AddProduct(ctx context.Context, role string, pvzId uuid.UUID, producttype string) (*models.Product, int, error) {
	args := m.Called(ctx, role, pvzId, producttype)
	var product *models.Product
	if args.Get(0) != nil {
		product = args.Get(0).(*models.Product)
	}
	return product, args.Int(1), args.Error(2)
}
func (m *MockService) GetPVZ(ctx context.Context) ([]*pb.PVZ, error) {
	return nil, nil
}

func TestDummyLoginHandler(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		handler := NewHandler(mockSvc)

		handler.DummyLoginHandler(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Неверный запрос", errResp.Message)
	})

	t.Run("service returns error", func(t *testing.T) {
		requestBody, _ := json.Marshal(models.DummyLoginRequest{Role: "invalid"})
		req := httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewBuffer(requestBody))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		mockSvc.On("DummyLogin", &models.DummyLoginRequest{Role: "invalid"}).
			Return("", http.StatusBadRequest, errors.New("неверная роль"))

		handler := NewHandler(mockSvc)
		handler.DummyLoginHandler(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "неверная роль", errResp.Message)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		reqBody, _ := json.Marshal(models.DummyLoginRequest{Role: "employee"})
		req := httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		expectedToken := "sometoken"
		mockSvc := new(MockService)
		mockSvc.On("DummyLogin", &models.DummyLoginRequest{Role: "employee"}).
			Return(expectedToken, http.StatusOK, nil)

		handler := NewHandler(mockSvc)
		handler.DummyLoginHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var token string
		err := json.NewDecoder(rr.Body).Decode(&token)
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
		mockSvc.AssertExpectations(t)
	})
}

func TestRegisterHandler(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		handler := NewHandler(mockSvc)
		handler.RegisterHandler(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Неверный запрос", errResp.Message)
	})

	t.Run("service returns error", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Email:    "user@example.com",
			Password: "pass",
			Role:     "invalid",
		}
		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)

		mockSvc.On("Register", mock.Anything, &registerReq).
			Return((*models.User)(nil), http.StatusBadRequest, errors.New("неверная роль"))

		handler := NewHandler(mockSvc)
		handler.RegisterHandler(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "неверная роль", errResp.Message)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		registerReq := models.RegisterRequest{
			Email:    "user@example.com",
			Password: "pass",
			Role:     "employee",
		}
		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		expectedUser := &models.User{
			ID:    uuid.New(),
			Email: "user@example.com",
			Role:  "employee",
		}
		mockSvc := new(MockService)
		mockSvc.On("Register", mock.Anything, &registerReq).
			Return(expectedUser, http.StatusOK, nil)

		handler := NewHandler(mockSvc)
		handler.RegisterHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var userResp models.User
		err := json.NewDecoder(rr.Body).Decode(&userResp)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, userResp.ID)
		assert.Equal(t, expectedUser.Email, userResp.Email)
		assert.Equal(t, expectedUser.Role, userResp.Role)
		mockSvc.AssertExpectations(t)
	})
}

func TestLoginHandler(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		handler := NewHandler(mockSvc)
		handler.LoginHandler(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Неверный запрос", errResp.Message)
	})

	t.Run("service returns error", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "user@example.com",
			Password: "wrongpass",
		}
		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		mockSvc.On("Login", mock.Anything, &loginReq).
			Return("", http.StatusUnauthorized, errors.New("неверные учетные данные"))

		handler := NewHandler(mockSvc)
		handler.LoginHandler(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "неверные учетные данные", errResp.Message)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "user@example.com",
			Password: "correctpass",
		}
		reqBody, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		expectedToken := "validtoken"
		mockSvc := new(MockService)
		mockSvc.On("Login", mock.Anything, &loginReq).
			Return(expectedToken, http.StatusOK, nil)

		handler := NewHandler(mockSvc)
		handler.LoginHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var token string
		err := json.NewDecoder(rr.Body).Decode(&token)
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
		mockSvc.AssertExpectations(t)
	})
}

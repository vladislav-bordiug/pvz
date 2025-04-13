package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"pvz/internal/models"
	pb "pvz/internal/pb/pvz_v1"
)

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		user.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockDatabase) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	var u *models.User
	if args.Get(0) != nil {
		u = args.Get(0).(*models.User)
	}
	return u, args.Error(1)
}

func (m *MockDatabase) CreatePVZ(ctx context.Context, pvz *models.PVZ) error {
	args := m.Called(ctx, pvz)
	if args.Error(0) == nil {
		pvz.ID = uuid.New()
	}
	return args.Error(0)
}
func (m *MockDatabase) GetPVZs(ctx context.Context, limit, offset int) ([]models.PVZ, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) != nil {
		return args.Get(0).([]models.PVZ), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockDatabase) GetReceptionsByPVZ(ctx context.Context, pvzId uuid.UUID, startDate, endDate *time.Time) ([]models.Reception, error) {
	args := m.Called(ctx, pvzId, startDate, endDate)
	if args.Get(0) != nil {
		return args.Get(0).([]models.Reception), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockDatabase) GetProductsByReception(ctx context.Context, receptionID uuid.UUID) ([]*models.Product, error) {
	args := m.Called(ctx, receptionID)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockDatabase) CloseLastReception(ctx context.Context, pvzId uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzId)
	if rec, ok := args.Get(0).(*models.Reception); ok {
		return rec, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockDatabase) DeleteLastProduct(ctx context.Context, pvzId uuid.UUID) error {
	args := m.Called(ctx, pvzId)
	return args.Error(0)
}
func (m *MockDatabase) CreateReception(ctx context.Context, pvzId uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzId)
	if rec, ok := args.Get(0).(*models.Reception); ok {
		return rec, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockDatabase) AddProduct(ctx context.Context, pvzId uuid.UUID, productType string) (*models.Product, error) {
	args := m.Called(ctx, pvzId, productType)
	if prod, ok := args.Get(0).(*models.Product); ok {
		return prod, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDatabase) GetPVZ(ctx context.Context) ([]*pb.PVZ, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]*pb.PVZ), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestDummyLogin(t *testing.T) {
	jwtSecret := []byte("testsecret")

	svc := NewService(nil, jwtSecret)

	t.Run("invalid role", func(t *testing.T) {
		token, status, err := svc.DummyLogin(&models.DummyLoginRequest{Role: "admin"})
		assert.Empty(t, token)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.EqualError(t, err, "неверная роль")
	})

	t.Run("valid role", func(t *testing.T) {
		token, status, err := svc.DummyLogin(&models.DummyLoginRequest{Role: "employee"})
		assert.NotEmpty(t, token)
		assert.GreaterOrEqual(t, strings.Count(token, "."), 2, "token should be a JWT")
		assert.Equal(t, http.StatusOK, status)
		assert.NoError(t, err)
	})
}

func TestRegister(t *testing.T) {
	jwtSecret := []byte("testsecret")
	mockDB := new(MockDatabase)
	svc := NewService(mockDB, jwtSecret)
	ctx := context.Background()

	t.Run("invalid role", func(t *testing.T) {
		user, status, err := svc.Register(ctx, &models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Role:     "admin",
		})
		assert.Nil(t, user)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.EqualError(t, err, "неверная роль")
	})

	t.Run("database create user error", func(t *testing.T) {
		req := &models.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Role:     "employee",
		}
		mockDB.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("db error")).Once()

		user, status, err := svc.Register(ctx, req)
		assert.Nil(t, user)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.EqualError(t, err, "ошибка регистрации: db error")
		mockDB.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		req := &models.RegisterRequest{
			Email:    "success@example.com",
			Password: "password123",
			Role:     "employee",
		}
		mockDB.On("CreateUser", ctx, mock.MatchedBy(func(user *models.User) bool {
			return user.Email == req.Email && user.Role == req.Role && user.Password != req.Password
		})).Return(nil).Once()

		user, status, err := svc.Register(ctx, req)
		assert.NotNil(t, user)
		assert.Equal(t, http.StatusOK, status)
		assert.NoError(t, err)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Role, user.Role)
		assert.NotEqual(t, uuid.Nil, user.ID)
		mockDB.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	jwtSecret := []byte("testsecret")
	mockDB := new(MockDatabase)
	svc := NewService(mockDB, jwtSecret)
	ctx := context.Background()

	plainPassword := "mysecretpass"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:       uuid.New(),
		Email:    "login@example.com",
		Password: string(hashedPassword),
		Role:     "employee",
	}

	t.Run("user not found", func(t *testing.T) {
		mockDB.On("GetUserByEmail", ctx, "notfound@example.com").Return((*models.User)(nil), errors.New("not found")).Once()

		token, status, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "notfound@example.com",
			Password: plainPassword,
		})
		assert.Empty(t, token)
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.EqualError(t, err, "неверные учетные данные")
		mockDB.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockDB.On("GetUserByEmail", ctx, "login@example.com").Return(testUser, nil).Once()

		token, status, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "login@example.com",
			Password: "wrongpassword",
		})
		assert.Empty(t, token)
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.EqualError(t, err, "неверные учетные данные")
		mockDB.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.On("GetUserByEmail", ctx, "login@example.com").Return(testUser, nil).Once()

		token, status, err := svc.Login(ctx, &models.LoginRequest{
			Email:    "login@example.com",
			Password: plainPassword,
		})
		assert.NotEmpty(t, token)
		assert.Equal(t, http.StatusOK, status)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, strings.Count(token, "."), 2, "token should be a JWT")
		mockDB.AssertExpectations(t)
	})
}

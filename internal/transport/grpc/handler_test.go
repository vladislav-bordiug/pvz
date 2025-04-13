package grpch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pvz/internal/models"
	pb "pvz/internal/pb/pvz_v1"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetPVZ(ctx context.Context) ([]*pb.PVZ, error) {
	args := m.Called(ctx)

	return args.Get(0).([]*pb.PVZ), args.Error(1)
}

func (m *MockService) DummyLogin(req *models.DummyLoginRequest) (string, int, error) {
	return "", 0, nil
}

func (m *MockService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, int, error) {
	return nil, 0, nil
}

func (m *MockService) Login(ctx context.Context, req *models.LoginRequest) (string, int, error) {
	return "", 0, nil
}

func (m *MockService) CreatePVZ(ctx context.Context, pvz *models.PVZ, role string) (int, error) {
	return 0, nil
}

func (m *MockService) ListPVZ(ctx context.Context, startDateStr, endDateStr, pageStr, limitStr string) ([]*models.PVZResponse, int, error) {
	return nil, 0, nil
}

func (m *MockService) CloseLastReception(ctx context.Context, role string, pvzId uuid.UUID) (*models.Reception, int, error) {
	return nil, 0, nil
}

func (m *MockService) DeleteLastProduct(ctx context.Context, role string, pvzId uuid.UUID) (int, error) {
	return 0, nil
}

func (m *MockService) CreateReception(ctx context.Context, role string, pvzId uuid.UUID) (*models.Reception, int, error) {
	return nil, 0, nil
}

func (m *MockService) AddProduct(ctx context.Context, role string, pvzId uuid.UUID, producttype string) (*models.Product, int, error) {
	return nil, 0, nil
}

func TestGetPVZList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockService)
		expectedPVZ := &pb.PVZ{
			Id:               "1",
			City:             "Moscow",
			RegistrationDate: timestamppb.New(time.Now()),
		}

		mockSvc.On("GetPVZ", mock.Anything).Return([]*pb.PVZ{expectedPVZ}, nil)

		server := NewGrpcServer(mockSvc)
		req := &pb.GetPVZListRequest{}

		resp, err := server.GetPVZList(context.Background(), req)
		assert.NoError(t, err, "Ожидалась ошибка nil")
		assert.NotNil(t, resp, "Ожидался ненулевой ответ")
		assert.Len(t, resp.Pvzs, 1, "Ожидался один элемент в списке")
		assert.Equal(t, expectedPVZ.Id, resp.Pvzs[0].Id, "Неверный Id в ответе")
		assert.Equal(t, expectedPVZ.City, resp.Pvzs[0].City, "Неверный город в ответе")

		mockSvc.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := new(MockService)
		errMessage := "ошибка получения ПВЗ"
		mockSvc.On("GetPVZ", mock.Anything).Return(([]*pb.PVZ)(nil), errors.New(errMessage))

		server := NewGrpcServer(mockSvc)
		req := &pb.GetPVZListRequest{}

		resp, err := server.GetPVZList(context.Background(), req)
		assert.Nil(t, resp, "В случае ошибки ответ должен быть nil")
		assert.EqualError(t, err, errMessage, "Неверное сообщение об ошибке")

		mockSvc.AssertExpectations(t)
	})
}

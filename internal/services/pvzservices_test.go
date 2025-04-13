package services

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pvz/internal/models"
	pb "pvz/internal/pb/pvz_v1"
)

func TestCreatePVZ(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		role           string
		inputPVZ       models.PVZ
		mockSetup      func(mdb *MockDatabase, pvz *models.PVZ)
		expectedStatus int
		expectedErr    string
	}{
		{
			name:     "role not moderator",
			role:     "employee",
			inputPVZ: models.PVZ{City: "Москва"},
			mockSetup: func(mdb *MockDatabase, pvz *models.PVZ) {
			},
			expectedStatus: http.StatusForbidden,
			expectedErr:    "доступ запрещен",
		},
		{
			name:     "invalid city",
			role:     "moderator",
			inputPVZ: models.PVZ{City: "Ростов"},
			mockSetup: func(mdb *MockDatabase, pvz *models.PVZ) {
			},
			expectedStatus: http.StatusBadRequest,

			expectedErr: "ошибка ПВЗ можно создать только",
		},
		{
			name:     "db create error",
			role:     "moderator",
			inputPVZ: models.PVZ{City: "Москва"},
			mockSetup: func(mdb *MockDatabase, pvz *models.PVZ) {
				mdb.On("CreatePVZ", ctx, pvz).Return(errors.New("db create error")).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedErr:    "ошибка создания ПВЗ: db create error",
		},
		{
			name:     "success",
			role:     "moderator",
			inputPVZ: models.PVZ{City: "Москва"},
			mockSetup: func(mdb *MockDatabase, pvz *models.PVZ) {
				mdb.On("CreatePVZ", ctx, pvz).Return(nil).Once().Run(func(args mock.Arguments) {
					p := args.Get(1).(*models.PVZ)
					p.ID = uuid.New()
					p.RegistrationDate = time.Now()
				})
			},
			expectedStatus: http.StatusOK,
			expectedErr:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.inputPVZ
			mockDB := new(MockDatabase)
			tt.mockSetup(mockDB, &input)

			svc := NewService(mockDB, []byte("unused"))
			status, err := svc.CreatePVZ(ctx, &input, tt.role)
			assert.Equal(t, tt.expectedStatus, status)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.False(t, input.RegistrationDate.IsZero())
				assert.NotEqual(t, uuid.Nil, input.ID)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestListPVZ(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		startDateStr      string
		endDateStr        string
		pageStr           string
		limitStr          string
		mockSetup         func(mdb *MockDatabase)
		expectedStatus    int
		expectedErrSubstr string
		expectedResults   int
	}{
		{
			name:         "error getting PVZs",
			startDateStr: "",
			endDateStr:   "",
			pageStr:      "1",
			limitStr:     "10",
			mockSetup: func(mdb *MockDatabase) {
				mdb.On("GetPVZs", ctx, 10, 0).Return(nil, errors.New("db get error")).Once()
			},
			expectedStatus:    http.StatusBadRequest,
			expectedErrSubstr: "ошибка выборки ПВЗ",
		},
		{
			name:         "error getting receptions",
			startDateStr: "",
			endDateStr:   "",
			pageStr:      "1",
			limitStr:     "10",
			mockSetup: func(mdb *MockDatabase) {

				pvz := models.PVZ{ID: uuid.New(), City: "Москва", RegistrationDate: time.Now()}
				mdb.On("GetPVZs", ctx, 10, 0).Return([]models.PVZ{pvz}, nil).Once()

				mdb.On("GetReceptionsByPVZ", ctx, pvz.ID, (*time.Time)(nil), (*time.Time)(nil)).Return(nil, errors.New("db receptions error")).Once()
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedErrSubstr: "ошибка выборки приёмок",
		},
		{
			name:         "error getting products",
			startDateStr: "",
			endDateStr:   "",
			pageStr:      "1",
			limitStr:     "10",
			mockSetup: func(mdb *MockDatabase) {

				pvz := models.PVZ{ID: uuid.New(), City: "Москва", RegistrationDate: time.Now()}
				mdb.On("GetPVZs", ctx, 10, 0).Return([]models.PVZ{pvz}, nil).Once()

				reception := models.Reception{ID: uuid.New(), PVZId: pvz.ID, Status: "in_progress", DateTime: time.Now()}
				mdb.On("GetReceptionsByPVZ", ctx, pvz.ID, (*time.Time)(nil), (*time.Time)(nil)).Return([]models.Reception{reception}, nil).Once()

				mdb.On("GetProductsByReception", ctx, reception.ID).Return(nil, errors.New("db products error")).Once()
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedErrSubstr: "ошибка выборки товаров",
		},
		{
			name:         "success",
			startDateStr: "2023-01-01T00:00:00Z",
			endDateStr:   "2023-01-02T00:00:00Z",
			pageStr:      "2",
			limitStr:     "5",
			mockSetup: func(mdb *MockDatabase) {
				pvz := models.PVZ{ID: uuid.New(), City: "Москва", RegistrationDate: time.Now()}
				mdb.On("GetPVZs", ctx, 5, 5).Return([]models.PVZ{pvz}, nil).Once()

				startDate, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00Z")
				mdb.On("GetReceptionsByPVZ", ctx, pvz.ID, &startDate, &endDate).Return([]models.Reception{}, nil).Once()
			},
			expectedStatus:  http.StatusOK,
			expectedResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDatabase)
			tt.mockSetup(mockDB)
			svc := NewService(mockDB, []byte("unused"))
			results, status, err := svc.ListPVZ(ctx, tt.startDateStr, tt.endDateStr, tt.pageStr, tt.limitStr)
			assert.Equal(t, tt.expectedStatus, status)
			if tt.expectedErrSubstr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrSubstr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.expectedResults)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetPVZ(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		mockSetup      func(mdb *MockDatabase)
		expectedErrSub string
		expectedCount  int
	}{
		{
			name: "db error",
			mockSetup: func(mdb *MockDatabase) {
				mdb.On("GetPVZ", ctx).Return(nil, errors.New("db getpvz error")).Once()
			},
			expectedErrSub: "ошибка получения ПВЗ: db getpvz error",
			expectedCount:  0,
		},
		{
			name: "success",
			mockSetup: func(mdb *MockDatabase) {
				pvz1 := &pb.PVZ{Id: "1", City: "Москва", RegistrationDate: timestamppb.New(time.Now())}
				pvz2 := &pb.PVZ{Id: "2", City: "Казань", RegistrationDate: timestamppb.New(time.Now())}
				mdb.On("GetPVZ", ctx).Return([]*pb.PVZ{pvz1, pvz2}, nil).Once()
			},
			expectedErrSub: "",
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDatabase)
			tt.mockSetup(mockDB)
			svc := NewService(mockDB, []byte("unused"))
			pvzs, err := svc.GetPVZ(ctx)
			if tt.expectedErrSub != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrSub)
				assert.Nil(t, pvzs)
			} else {
				assert.NoError(t, err)
				assert.Len(t, pvzs, tt.expectedCount)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

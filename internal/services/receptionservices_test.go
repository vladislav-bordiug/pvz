package services

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"pvz/internal/models"
)

func TestCloseLastReception(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()

	t.Run("role not employee", func(t *testing.T) {
		svc := NewService(nil, []byte("unused"))
		rec, status, err := svc.CloseLastReception(ctx, "moderator", pvzId)
		assert.Nil(t, rec)
		assert.Equal(t, http.StatusForbidden, status)
		assert.EqualError(t, err, "доступ запрещен")
	})

	t.Run("db error", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockDB.On("CloseLastReception", ctx, pvzId).Return(nil, errors.New("db error")).Once()

		svc := NewService(mockDB, []byte("unused"))
		rec, status, err := svc.CloseLastReception(ctx, "employee", pvzId)
		assert.Nil(t, rec)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.EqualError(t, err, "ошибка закрытия приёмки: db error")
		mockDB.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		expectedRec := &models.Reception{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZId:    pvzId,
			Status:   "close",
		}
		mockDB := new(MockDatabase)
		mockDB.On("CloseLastReception", ctx, pvzId).Return(expectedRec, nil).Once()

		svc := NewService(mockDB, []byte("unused"))
		rec, status, err := svc.CloseLastReception(ctx, "employee", pvzId)
		assert.Equal(t, http.StatusOK, status)
		assert.NoError(t, err)
		assert.Equal(t, expectedRec, rec)
		mockDB.AssertExpectations(t)
	})
}

func TestCreateReception(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()

	t.Run("role not employee", func(t *testing.T) {
		svc := NewService(nil, []byte("unused"))
		rec, status, err := svc.CreateReception(ctx, "moderator", pvzId)
		assert.Nil(t, rec)
		assert.Equal(t, http.StatusForbidden, status)
		assert.EqualError(t, err, "доступ запрещен")
	})

	t.Run("db error", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockDB.On("CreateReception", ctx, pvzId).Return(nil, errors.New("Активная приёмка уже существует")).Once()

		svc := NewService(mockDB, []byte("unused"))
		rec, status, err := svc.CreateReception(ctx, "employee", pvzId)
		assert.Nil(t, rec)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.EqualError(t, err, "ошибка создания приёмки: Активная приёмка уже существует")
		mockDB.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		expectedRec := &models.Reception{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZId:    pvzId,
			Status:   "in_progress",
		}
		mockDB := new(MockDatabase)
		mockDB.On("CreateReception", ctx, pvzId).Return(expectedRec, nil).Once()

		svc := NewService(mockDB, []byte("unused"))
		rec, status, err := svc.CreateReception(ctx, "employee", pvzId)
		assert.NotNil(t, rec)
		assert.Equal(t, http.StatusOK, status)
		assert.NoError(t, err)
		assert.Equal(t, expectedRec, rec)
		mockDB.AssertExpectations(t)
	})
}

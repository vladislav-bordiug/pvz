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

func TestDeleteLastProduct(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()

	t.Run("not employee", func(t *testing.T) {
		svc := NewService(nil, []byte("unused"))
		status, err := svc.DeleteLastProduct(ctx, "moderator", pvzId)
		assert.Equal(t, http.StatusForbidden, status)
		assert.EqualError(t, err, "доступ запрещен")
	})

	t.Run("db error", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockDB.On("DeleteLastProduct", ctx, pvzId).Return(errors.New("db error")).Once()
		svc := NewService(mockDB, []byte("unused"))
		status, err := svc.DeleteLastProduct(ctx, "employee", pvzId)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.EqualError(t, err, "ошибка удаления товара: db error")
		mockDB.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockDB.On("DeleteLastProduct", ctx, pvzId).Return(nil).Once()
		svc := NewService(mockDB, []byte("unused"))
		status, err := svc.DeleteLastProduct(ctx, "employee", pvzId)
		assert.Equal(t, http.StatusOK, status)
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})
}

func TestAddProduct(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()
	productType := "testType"

	t.Run("not employee", func(t *testing.T) {
		svc := NewService(nil, []byte("unused"))
		prod, status, err := svc.AddProduct(ctx, "moderator", pvzId, productType)
		assert.Nil(t, prod)
		assert.Equal(t, http.StatusForbidden, status)
		assert.EqualError(t, err, "доступ запрещен")
	})

	t.Run("db error", func(t *testing.T) {
		mockDB := new(MockDatabase)
		mockDB.On("AddProduct", ctx, pvzId, productType).Return(nil, errors.New("db add error")).Once()
		svc := NewService(mockDB, []byte("unused"))
		prod, status, err := svc.AddProduct(ctx, "employee", pvzId, productType)
		assert.Nil(t, prod)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.EqualError(t, err, "ошибка добавления товара: db add error")
		mockDB.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		expectedProduct := &models.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        productType,
			ReceptionId: uuid.New(),
		}
		mockDB := new(MockDatabase)
		mockDB.On("AddProduct", ctx, pvzId, productType).Return(expectedProduct, nil).Once()
		svc := NewService(mockDB, []byte("unused"))
		prod, status, err := svc.AddProduct(ctx, "employee", pvzId, productType)
		assert.NotNil(t, prod)
		assert.Equal(t, http.StatusOK, status)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.ID, prod.ID)
		assert.Equal(t, expectedProduct.Type, prod.Type)
		mockDB.AssertExpectations(t)
	})
}

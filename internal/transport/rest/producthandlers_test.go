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

func TestDeleteLastProductHandler(t *testing.T) {
	t.Run("invalid pvz id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/delete-product", nil)
		req = mux.SetURLVars(req, map[string]string{"pvzId": "invalid-uuid"})

		req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, "employee"))

		rr := httptest.NewRecorder()
		mockSvc := new(MockService)
		handler := NewHandler(mockSvc)

		handler.DeleteLastProductHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Неверный идентификатор ПВЗ", errResp.Message)
	})

	t.Run("service returns error", func(t *testing.T) {
		validUUID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/delete-product", nil)
		req = mux.SetURLVars(req, map[string]string{"pvzId": validUUID.String()})
		req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, "employee"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		errMsg := "ошибка удаления товара: some error"
		mockSvc.
			On("DeleteLastProduct", mock.Anything, "employee", validUUID).
			Return(http.StatusBadRequest, errors.New(errMsg))

		handler := NewHandler(mockSvc)
		handler.DeleteLastProductHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, errMsg, errResp.Message)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		validUUID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/delete-product", nil)
		req = mux.SetURLVars(req, map[string]string{"pvzId": validUUID.String()})
		req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, "employee"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		mockSvc.
			On("DeleteLastProduct", mock.Anything, "employee", validUUID).
			Return(http.StatusOK, nil)

		handler := NewHandler(mockSvc)
		handler.DeleteLastProductHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp map[string]string
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "Товар удалён", resp["message"])
		mockSvc.AssertExpectations(t)
	})
}
func TestAddProductHandler(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/add-product", bytes.NewBufferString("invalid json"))
		req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, "employee"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		handler := NewHandler(mockSvc)
		handler.AddProductHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Неверный запрос", errResp.Message)
	})

	t.Run("invalid pvz id", func(t *testing.T) {
		reqBody := `{"PVZId": "invalid-uuid", "type": "someType"}`
		req := httptest.NewRequest(http.MethodPost, "/add-product", bytes.NewBufferString(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, "employee"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		handler := NewHandler(mockSvc)
		handler.AddProductHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Неверный идентификатор ПВЗ", errResp.Message)
	})

	t.Run("service returns error", func(t *testing.T) {
		validUUID := uuid.New()
		reqData := models.AddProductRequest{
			PVZId: validUUID.String(),
			Type:  "someType",
		}
		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/add-product", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, "employee"))
		rr := httptest.NewRecorder()

		mockSvc := new(MockService)
		errMsg := "ошибка добавления товара: some error"
		mockSvc.
			On("AddProduct", mock.Anything, "employee", validUUID, "someType").
			Return((*models.Product)(nil), http.StatusBadRequest, errors.New(errMsg))

		handler := NewHandler(mockSvc)
		handler.AddProductHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp models.ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Equal(t, errMsg, errResp.Message)
		mockSvc.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		validUUID := uuid.New()
		reqData := models.AddProductRequest{
			PVZId: validUUID.String(),
			Type:  "someType",
		}
		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/add-product", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), contextkeys.ContextKeyRole, "employee"))
		rr := httptest.NewRecorder()

		expectedProduct := &models.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			Type:        "someType",
			ReceptionId: uuid.New(),
		}

		mockSvc := new(MockService)
		mockSvc.
			On("AddProduct", mock.Anything, "employee", validUUID, "someType").
			Return(expectedProduct, http.StatusOK, nil)

		handler := NewHandler(mockSvc)
		handler.AddProductHandler(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var prod models.Product
		err := json.NewDecoder(rr.Body).Decode(&prod)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.ID, prod.ID)
		assert.Equal(t, expectedProduct.Type, prod.Type)

		mockSvc.AssertExpectations(t)
	})
}

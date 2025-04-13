package rest

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"pvz/internal/contextkeys"
	"pvz/internal/metrics"
	"pvz/internal/models"
)

func (h *Handler) DeleteLastProductHandler(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(contextkeys.ContextKeyRole).(string)
	vars := mux.Vars(r)
	pvzIdStr := vars["pvzId"]
	pvzId, err := uuid.Parse(pvzIdStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный идентификатор ПВЗ"})
		logrus.WithError(err).Error("Ошибка DeleteLastProduct")
		return
	}
	status, err := h.services.DeleteLastProduct(r.Context(), role, pvzId)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка DeleteLastProduct")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("DeleteLastProduct выполнен успешно")
	json.NewEncoder(w).Encode(map[string]string{"message": "Товар удалён"})
}

func (h *Handler) AddProductHandler(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(contextkeys.ContextKeyRole).(string)
	var req models.AddProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный запрос"})
		logrus.WithError(err).Error("Ошибка AddProduct")
		return
	}
	pvzId, err := uuid.Parse(req.PVZId)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный идентификатор ПВЗ"})
		logrus.WithError(err).Error("Ошибка AddProduct")
		return
	}
	product, status, err := h.services.AddProduct(r.Context(), role, pvzId, req.Type)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка AddProduct")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("AddProduct выполнен успешно")
	metrics.AddedProductsTotal.Inc()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

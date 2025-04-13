package rest

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"pvz/internal/contextkeys"
	"pvz/internal/metrics"
	"pvz/internal/models"
)

func (h *Handler) CloseLastReceptionHandler(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(contextkeys.ContextKeyRole).(string)
	vars := mux.Vars(r)
	pvzIdStr := vars["pvzId"]
	pvzId, err := uuid.Parse(pvzIdStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный идентификатор ПВЗ"})
		logrus.WithError(err).Error("Ошибка CloseLastReception")
		return
	}
	rec, status, err := h.services.CloseLastReception(r.Context(), role, pvzId)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка CloseLastReception")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("CloseLastReception выполнен успешно")
	json.NewEncoder(w).Encode(rec)
}

func (h *Handler) CreateReceptionHandler(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(contextkeys.ContextKeyRole).(string)
	var req models.CreateReceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный запрос"})
		logrus.WithError(err).Error("Ошибка CreateReception")
		return
	}
	pvzId, err := uuid.Parse(req.PVZId)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный идентификатор ПВЗ"})
		logrus.WithError(err).Error("Ошибка CreateReception")
		return
	}
	rec, status, err := h.services.CreateReception(r.Context(), role, pvzId)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка CreateReception")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("CreateReception выполнен успешно")
	metrics.CreatedReceptionTotal.Inc()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rec)
}

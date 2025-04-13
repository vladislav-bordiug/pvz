package rest

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"pvz/internal/contextkeys"
	"pvz/internal/metrics"
	"pvz/internal/models"
)

func (h *Handler) CreatePVZHandler(w http.ResponseWriter, r *http.Request) {
	var pvz models.PVZ
	if err := json.NewDecoder(r.Body).Decode(&pvz); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный запрос"})
		logrus.WithError(err).Error("Ошибка CreatePVZ")
		return
	}
	role := r.Context().Value(contextkeys.ContextKeyRole).(string)
	status, err := h.services.CreatePVZ(r.Context(), &pvz, role)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка CreatePVZ")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("CreatePVZ выполнен успешно")
	metrics.CreatedPVZTotal.Inc()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pvz)
}

func (h *Handler) ListPVZHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	startDateStr := q.Get("startDate")
	endDateStr := q.Get("endDate")
	pageStr := q.Get("page")
	limitStr := q.Get("limit")
	results, status, err := h.services.ListPVZ(r.Context(), startDateStr, endDateStr, pageStr, limitStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка ListPVZ")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("ListPVZ выполнен успешно")
	json.NewEncoder(w).Encode(results)
}

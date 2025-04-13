package rest

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"pvz/internal/models"
	"pvz/internal/services"
)

type Handler struct {
	services services.ServiceInterface
}

func NewHandler(services services.ServiceInterface) *Handler {
	return &Handler{services: services}
}

func (h *Handler) DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный запрос"})
		logrus.WithError(err).Error("Ошибка DummyLogin")
		return
	}
	token, status, err := h.services.DummyLogin(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка DummyLogin")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("DummyLogin выполнен успешно")
	json.NewEncoder(w).Encode(token)
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный запрос"})
		logrus.WithError(err).Error("Ошибка Register")
		return
	}
	user, status, err := h.services.Register(r.Context(), &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка Register")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("Register выполнен успешно")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Неверный запрос"})
		logrus.WithError(err).Error("Ошибка Login")
		return
	}
	token, status, err := h.services.Login(r.Context(), &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
		logrus.WithError(err).Error("Ошибка Login")
		return
	}
	logrus.WithFields(logrus.Fields{
		"status": status,
	}).Info("Login выполнен успешно")
	json.NewEncoder(w).Encode(token)
}

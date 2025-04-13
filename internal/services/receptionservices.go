package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"pvz/internal/models"
)

func (s *Service) CloseLastReception(ctx context.Context, role string, pvzId uuid.UUID) (rec *models.Reception, status int, err error) {
	if role != "employee" {
		return rec, http.StatusForbidden, errors.New("доступ запрещен")
	}
	rec, err = s.database.CloseLastReception(ctx, pvzId)
	if err != nil {
		return rec, http.StatusBadRequest, errors.New("ошибка закрытия приёмки: " + err.Error())
	}
	return rec, http.StatusOK, nil
}

func (s *Service) CreateReception(ctx context.Context, role string, pvzId uuid.UUID) (rec *models.Reception, status int, err error) {
	if role != "employee" {
		return rec, http.StatusForbidden, errors.New("доступ запрещен")
	}
	rec, err = s.database.CreateReception(ctx, pvzId)
	if err != nil {
		return rec, http.StatusBadRequest, errors.New("ошибка создания приёмки: " + err.Error())
	}
	return rec, http.StatusOK, nil
}

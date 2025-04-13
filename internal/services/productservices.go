package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"pvz/internal/models"
)

func (s *Service) DeleteLastProduct(ctx context.Context, role string, pvzId uuid.UUID) (status int, err error) {
	if role != "employee" {
		return http.StatusForbidden, errors.New("доступ запрещен")
	}
	err = s.database.DeleteLastProduct(ctx, pvzId)
	if err != nil {
		return http.StatusBadRequest, errors.New("ошибка удаления товара: " + err.Error())
	}
	return http.StatusOK, nil
}

func (s *Service) AddProduct(ctx context.Context, role string, pvzId uuid.UUID, producttype string) (product *models.Product, status int, err error) {
	if role != "employee" {
		return product, http.StatusForbidden, errors.New("доступ запрещен")
	}
	product, err = s.database.AddProduct(ctx, pvzId, producttype)
	if err != nil {
		return product, http.StatusBadRequest, errors.New("ошибка добавления товара: " + err.Error())
	}
	return product, http.StatusOK, nil
}

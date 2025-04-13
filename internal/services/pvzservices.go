package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"pvz/internal/models"
	pb "pvz/internal/pb/pvz_v1"
)

func (s *Service) CreatePVZ(ctx context.Context, pvz *models.PVZ, role string) (status int, err error) {
	if role != "moderator" {
		return http.StatusForbidden, errors.New("доступ запрещен")
	}
	if pvz.City != "Москва" && pvz.City != "Санкт-Петербург" && pvz.City != "Казань" {
		return http.StatusBadRequest, errors.New(fmt.Sprintf("ошибка ПВЗ можно создать только в Москве, Санкт-Петербурге или Казани %v", err))
	}
	pvz.RegistrationDate = time.Now()
	if err := s.database.CreatePVZ(ctx, pvz); err != nil {
		return http.StatusBadRequest, errors.New(fmt.Sprintf("ошибка создания ПВЗ: %v", err))
	}
	return http.StatusOK, nil
}

func (s *Service) ListPVZ(ctx context.Context, startDateStr string, endDateStr string, pageStr string, limitStr string) (results []*models.PVZResponse, status int, err error) {
	page := 1
	limit := 10
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l >= 1 && l <= 30 {
			limit = l
		}
	}
	offset := (page - 1) * limit
	pvzs, err := s.database.GetPVZs(ctx, limit, offset)
	if err != nil {
		return results, http.StatusBadRequest, errors.New("ошибка выборки ПВЗ")
	}
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &t
		}
	}
	for _, pvz := range pvzs {
		recs, err := s.database.GetReceptionsByPVZ(ctx, pvz.ID, startDate, endDate)
		if err != nil {
			return results, http.StatusInternalServerError, errors.New("ошибка выборки приёмок")
		}
		var recInfos []*models.ReceptionInfo
		for _, rec := range recs {
			products, err := s.database.GetProductsByReception(ctx, rec.ID)
			if err != nil {
				return results, http.StatusInternalServerError, errors.New("ошибка выборки товаров")
			}
			recInfos = append(recInfos, &models.ReceptionInfo{
				Reception: &rec,
				Products:  products,
			})
		}
		results = append(results, &models.PVZResponse{
			PVZ:        &pvz,
			Receptions: recInfos,
		})
	}
	return results, http.StatusOK, nil
}

func (s *Service) GetPVZ(ctx context.Context) (pvzs []*pb.PVZ, err error) {
	pvzs, err = s.database.GetPVZ(ctx)
	if err != nil {
		return pvzs, errors.New("ошибка получения ПВЗ: " + err.Error())
	}
	return pvzs, nil
}

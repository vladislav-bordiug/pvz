package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pvz/internal/models"
	pb "pvz/internal/pb/pvz_v1"
)

func (db *PGXDatabase) CreatePVZ(ctx context.Context, pvz *models.PVZ) (err error) {
	query := `INSERT INTO pvz (registration_date, city) VALUES ($1, $2) RETURNING id`
	return db.pool.QueryRow(ctx, query, pvz.RegistrationDate, pvz.City).Scan(&pvz.ID)
}

func (db *PGXDatabase) GetPVZs(ctx context.Context, limit, offset int) (pvzs []models.PVZ, err error) {
	query := `SELECT id, registration_date, city FROM pvz ORDER BY registration_date DESC LIMIT $1 OFFSET $2`
	rows, err := db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.PVZ
		if err := rows.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
			return nil, err
		}
		pvzs = append(pvzs, p)
	}
	return pvzs, nil
}

func (db *PGXDatabase) GetReceptionsByPVZ(ctx context.Context, pvzId uuid.UUID, startDate, endDate *time.Time) (recs []models.Reception, err error) {
	query := `SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id=$1`
	args := []interface{}{pvzId}

	conditions := []string{}
	if startDate != nil {
		conditions = append(conditions, fmt.Sprintf("date_time >= $%d", len(args)+1))
		args = append(args, *startDate)
	}
	if endDate != nil {
		conditions = append(conditions, fmt.Sprintf("date_time <= $%d", len(args)+1))
		args = append(args, *endDate)
	}
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY date_time DESC"

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rec models.Reception
		if err := rows.Scan(&rec.ID, &rec.DateTime, &rec.PVZId, &rec.Status); err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func (db *PGXDatabase) GetProductsByReception(ctx context.Context, receptionID uuid.UUID) (products []*models.Product, err error) {
	query := `SELECT id, date_time, type, reception_id FROM products WHERE reception_id=$1 ORDER BY date_time ASC`
	rows, err := db.pool.Query(ctx, query, receptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var prod models.Product
		if err := rows.Scan(&prod.ID, &prod.DateTime, &prod.Type, &prod.ReceptionId); err != nil {
			return nil, err
		}
		products = append(products, &prod)
	}
	return products, nil
}

func (db *PGXDatabase) GetPVZ(ctx context.Context) (pvzs []*pb.PVZ, err error) {
	query := `SELECT id, registration_date, city FROM pvz`
	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var regDate time.Time
		var city string

		if err := rows.Scan(&id, &regDate, &city); err != nil {
			return nil, err
		}

		pvzs = append(pvzs, &pb.PVZ{
			Id:               id,
			RegistrationDate: timestamppb.New(regDate),
			City:             city,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pvzs, nil
}

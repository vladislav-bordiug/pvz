package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"pvz/internal/models"
)

func (db *PGXDatabase) CloseLastReception(ctx context.Context, pvzId uuid.UUID) (rec *models.Reception, err error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return rec, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	query := `
		SELECT id, date_time, pvz_id, status
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1
	`
	rec = &models.Reception{}
	err = tx.QueryRow(ctx, query, pvzId).Scan(&rec.ID, &rec.DateTime, &rec.PVZId, &rec.Status)
	if err != nil {
		tx.Rollback(ctx)
		return rec, err
	}
	updateQuery := `
		UPDATE receptions
		SET status='close'
		WHERE id=$1
		RETURNING id, date_time, pvz_id, status
	`
	err = tx.QueryRow(ctx, updateQuery, rec.ID).Scan(&rec.ID, &rec.DateTime, &rec.PVZId, &rec.Status)
	if err != nil {
		tx.Rollback(ctx)
		return rec, err
	}
	err = tx.Commit(ctx)
	return rec, err
}

func (db *PGXDatabase) CreateReception(ctx context.Context, pvzId uuid.UUID) (rec *models.Reception, err error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return rec, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	var count int
	checkQuery := `SELECT COUNT(*) FROM receptions WHERE pvz_id=$1 AND status='in_progress'`
	err = tx.QueryRow(ctx, checkQuery, pvzId).Scan(&count)
	if err != nil {
		tx.Rollback(ctx)
		return rec, err
	}
	if count > 0 {
		tx.Rollback(ctx)
		return rec, errors.New("Активная приёмка уже существует")
	}
	rec = &models.Reception{
		DateTime: time.Now(),
		PVZId:    pvzId,
		Status:   "in_progress",
	}
	insertQuery := `INSERT INTO receptions (date_time, pvz_id, status) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(ctx, insertQuery, rec.DateTime, rec.PVZId, rec.Status).Scan(&rec.ID)
	if err != nil {
		tx.Rollback(ctx)
		return rec, err
	}
	err = tx.Commit(ctx)
	return rec, err
}

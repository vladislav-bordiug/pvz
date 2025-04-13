package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"pvz/internal/models"
)

func (db *PGXDatabase) DeleteLastProduct(ctx context.Context, pvzId uuid.UUID) (err error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	var receptionID uuid.UUID
	query := `
		SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1
	`
	err = tx.QueryRow(ctx, query, pvzId).Scan(&receptionID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	var productID uuid.UUID
	productQuery := `
		SELECT id
		FROM products
		WHERE reception_id=$1
		ORDER BY date_time DESC
		LIMIT 1
	`
	err = tx.QueryRow(ctx, productQuery, receptionID).Scan(&productID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	deleteQuery := `DELETE FROM products WHERE id=$1`
	_, err = tx.Exec(ctx, deleteQuery, productID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (db *PGXDatabase) AddProduct(ctx context.Context, pvzId uuid.UUID, productType string) (product *models.Product, err error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return product, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	var receptionID uuid.UUID
	query := `
		SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1
	`
	err = tx.QueryRow(ctx, query, pvzId).Scan(&receptionID)
	if err != nil {
		tx.Rollback(ctx)
		return product, errors.New("Нет активной приёмки для данного ПВЗ")
	}
	product = &models.Product{
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionId: receptionID,
	}
	insertQuery := `INSERT INTO products (date_time, type, reception_id) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(ctx, insertQuery, product.DateTime, product.Type, product.ReceptionId).Scan(&product.ID)
	if err != nil {
		tx.Rollback(ctx)
		return product, err
	}
	err = tx.Commit(ctx)
	return product, err
}

package database

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestDeleteLastProduct(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()

	t.Run("begin error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		expectedErr := errors.New("begin error")
		mockPool.ExpectBegin().WillReturnError(expectedErr)

		db := NewPGXDatabase(mockPool)
		err = db.DeleteLastProduct(ctx, pvzId)
		assert.EqualError(t, err, expectedErr.Error())

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("reception query error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		expectedErr := errors.New("reception query error")
		mockPool.ExpectBegin()
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnError(expectedErr)
		mockPool.ExpectRollback()

		db := NewPGXDatabase(mockPool)
		err = db.DeleteLastProduct(ctx, pvzId)
		assert.EqualError(t, err, expectedErr.Error())

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("product query error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		receptionID := uuid.New()
		mockPool.ExpectBegin()
		rowsReception := pgxmock.NewRows([]string{"id"}).AddRow(receptionID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnRows(rowsReception)
		expectedErr := errors.New("product query error")
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM products
		WHERE reception_id=$1
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(receptionID).
			WillReturnError(expectedErr)
		mockPool.ExpectRollback()

		db := NewPGXDatabase(mockPool)
		err = db.DeleteLastProduct(ctx, pvzId)
		assert.EqualError(t, err, expectedErr.Error())

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		receptionID := uuid.New()
		productID := uuid.New()

		mockPool.ExpectBegin()
		rowsReception := pgxmock.NewRows([]string{"id"}).AddRow(receptionID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnRows(rowsReception)
		rowsProduct := pgxmock.NewRows([]string{"id"}).AddRow(productID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM products
		WHERE reception_id=$1
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(receptionID).
			WillReturnRows(rowsProduct)
		expectedErr := errors.New("delete exec error")
		mockPool.
			ExpectExec(regexp.QuoteMeta("DELETE FROM products WHERE id=$1")).
			WithArgs(productID).
			WillReturnError(expectedErr)
		mockPool.ExpectRollback()

		db := NewPGXDatabase(mockPool)
		err = db.DeleteLastProduct(ctx, pvzId)
		assert.EqualError(t, err, expectedErr.Error())

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		receptionID := uuid.New()
		productID := uuid.New()

		mockPool.ExpectBegin()
		rowsReception := pgxmock.NewRows([]string{"id"}).AddRow(receptionID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnRows(rowsReception)
		rowsProduct := pgxmock.NewRows([]string{"id"}).AddRow(productID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM products
		WHERE reception_id=$1
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(receptionID).
			WillReturnRows(rowsProduct)
		mockPool.
			ExpectExec(regexp.QuoteMeta("DELETE FROM products WHERE id=$1")).
			WithArgs(productID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		expectedErr := errors.New("commit error")
		mockPool.ExpectCommit().WillReturnError(expectedErr)

		db := NewPGXDatabase(mockPool)
		err = db.DeleteLastProduct(ctx, pvzId)
		assert.EqualError(t, err, expectedErr.Error())

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("success", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		receptionID := uuid.New()
		productID := uuid.New()

		mockPool.ExpectBegin()
		rowsReception := pgxmock.NewRows([]string{"id"}).AddRow(receptionID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnRows(rowsReception)
		rowsProduct := pgxmock.NewRows([]string{"id"}).AddRow(productID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM products
		WHERE reception_id=$1
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(receptionID).
			WillReturnRows(rowsProduct)
		mockPool.
			ExpectExec(regexp.QuoteMeta("DELETE FROM products WHERE id=$1")).
			WithArgs(productID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mockPool.ExpectCommit()

		db := NewPGXDatabase(mockPool)
		err = db.DeleteLastProduct(ctx, pvzId)
		assert.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestAddProduct(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()
	productType := "testProduct"

	t.Run("begin error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		expectedErr := errors.New("begin error")
		mockPool.ExpectBegin().WillReturnError(expectedErr)

		db := NewPGXDatabase(mockPool)
		product, err := db.AddProduct(ctx, pvzId, productType)
		assert.Nil(t, product)
		assert.EqualError(t, err, expectedErr.Error())
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("no active reception", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		mockPool.ExpectBegin()
		expectedErr := errors.New("query error")
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnError(expectedErr)
		mockPool.ExpectRollback()

		db := NewPGXDatabase(mockPool)
		product, err := db.AddProduct(ctx, pvzId, productType)
		assert.Nil(t, product)
		assert.EqualError(t, err, "Нет активной приёмки для данного ПВЗ")
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("insert error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		receptionID := uuid.New()
		mockPool.ExpectBegin()
		rowsReception := pgxmock.NewRows([]string{"id"}).AddRow(receptionID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnRows(rowsReception)
		expectedErr := errors.New("insert error")
		mockPool.
			ExpectQuery(regexp.QuoteMeta("INSERT INTO products (date_time, type, reception_id) VALUES ($1, $2, $3) RETURNING id")).
			WithArgs(pgxmock.AnyArg(), productType, receptionID).
			WillReturnError(expectedErr)
		mockPool.ExpectRollback()

		db := NewPGXDatabase(mockPool)
		product, err := db.AddProduct(ctx, pvzId, productType)
		assert.NotNil(t, product)
		assert.EqualError(t, err, expectedErr.Error())
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		receptionID := uuid.New()
		mockPool.ExpectBegin()
		rowsReception := pgxmock.NewRows([]string{"id"}).AddRow(receptionID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnRows(rowsReception)
		newProductID := uuid.New()
		rowsInsert := pgxmock.NewRows([]string{"id"}).AddRow(newProductID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta("INSERT INTO products (date_time, type, reception_id) VALUES ($1, $2, $3) RETURNING id")).
			WithArgs(pgxmock.AnyArg(), productType, receptionID).
			WillReturnRows(rowsInsert)
		expectedErr := errors.New("commit error")
		mockPool.ExpectCommit().WillReturnError(expectedErr)

		db := NewPGXDatabase(mockPool)
		product, err := db.AddProduct(ctx, pvzId, productType)

		assert.NotNil(t, product)
		assert.EqualError(t, err, expectedErr.Error())
		assert.Equal(t, newProductID, product.ID)
		assert.Equal(t, productType, product.Type)
		assert.Equal(t, receptionID, product.ReceptionId)
	})

	t.Run("success", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		receptionID := uuid.New()
		mockPool.ExpectBegin()
		rowsReception := pgxmock.NewRows([]string{"id"}).AddRow(receptionID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta(`SELECT id
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
			WithArgs(pvzId).
			WillReturnRows(rowsReception)
		newProductID := uuid.New()
		rowsInsert := pgxmock.NewRows([]string{"id"}).AddRow(newProductID.String())
		mockPool.
			ExpectQuery(regexp.QuoteMeta("INSERT INTO products (date_time, type, reception_id) VALUES ($1, $2, $3) RETURNING id")).
			WithArgs(pgxmock.AnyArg(), productType, receptionID).
			WillReturnRows(rowsInsert)
		mockPool.ExpectCommit()

		db := NewPGXDatabase(mockPool)
		product, err := db.AddProduct(ctx, pvzId, productType)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, newProductID, product.ID)
		assert.Equal(t, productType, product.Type)
		assert.Equal(t, receptionID, product.ReceptionId)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

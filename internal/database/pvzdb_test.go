package database

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"pvz/internal/models"
)

func TestCreatePVZ(t *testing.T) {
	ctx := context.Background()

	t.Run("CreatePVZ", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			db := NewPGXDatabase(mockPool)
			pvz := &models.PVZ{
				RegistrationDate: time.Now(),
				City:             "Москва",
			}
			expectedID := uuid.New()
			mockPool.
				ExpectQuery(regexp.QuoteMeta("INSERT INTO pvz (registration_date, city) VALUES ($1, $2) RETURNING id")).
				WithArgs(pvz.RegistrationDate, pvz.City).
				WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedID.String()))

			err = db.CreatePVZ(ctx, pvz)
			assert.NoError(t, err)
			assert.Equal(t, expectedID, pvz.ID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Query Error", func(t *testing.T) {

			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			db := NewPGXDatabase(mockPool)
			pvz := &models.PVZ{
				RegistrationDate: time.Now(),
				City:             "Москва",
			}
			expectedErr := errors.New("query error")
			mockPool.
				ExpectQuery(regexp.QuoteMeta("INSERT INTO pvz (registration_date, city) VALUES ($1, $2) RETURNING id")).
				WithArgs(pvz.RegistrationDate, pvz.City).
				WillReturnError(expectedErr)

			err = db.CreatePVZ(ctx, pvz)
			assert.EqualError(t, err, expectedErr.Error())

			assert.Equal(t, uuid.Nil, pvz.ID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	})
}

func TestGetPVZs(t *testing.T) {
	ctx := context.Background()

	t.Run("GetPVZs", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			db := NewPGXDatabase(mockPool)

			id1 := uuid.New()
			regDate1 := time.Now().Add(-1 * time.Hour)
			city1 := "Москва"

			id2 := uuid.New()
			regDate2 := time.Now().Add(-2 * time.Hour)
			city2 := "Казань"

			rows := pgxmock.NewRows([]string{"id", "registration_date", "city"}).
				AddRow(id1.String(), regDate1, city1).
				AddRow(id2.String(), regDate2, city2)

			limit, offset := 10, 0
			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT id, registration_date, city FROM pvz ORDER BY registration_date DESC LIMIT $1 OFFSET $2")).
				WithArgs(limit, offset).
				WillReturnRows(rows)

			pvzs, err := db.GetPVZs(ctx, limit, offset)
			assert.NoError(t, err)
			assert.Len(t, pvzs, 2)

			assert.Equal(t, id1, pvzs[0].ID)
			assert.Equal(t, city1, pvzs[0].City)

			assert.Equal(t, id2, pvzs[1].ID)
			assert.Equal(t, city2, pvzs[1].City)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Query Error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			db := NewPGXDatabase(mockPool)
			limit, offset := 10, 0
			expectedErr := errors.New("select error")
			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT id, registration_date, city FROM pvz ORDER BY registration_date DESC LIMIT $1 OFFSET $2")).
				WithArgs(limit, offset).
				WillReturnError(expectedErr)

			pvzs, err := db.GetPVZs(ctx, limit, offset)
			assert.Nil(t, pvzs)
			assert.EqualError(t, err, expectedErr.Error())

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	})
}

func TestGetReceptionsByPVZ(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()

	t.Run("GetReceptionsByPVZ", func(t *testing.T) {
		t.Run("Nil Dates", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			db := NewPGXDatabase(mockPool)
			receptionID := uuid.New()
			recDate := time.Now().Add(-30 * time.Minute)
			status := "in_progress"
			rows := pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
				AddRow(receptionID.String(), recDate, pvzId.String(), status)

			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id=$1 ORDER BY date_time DESC")).
				WithArgs(pvzId).
				WillReturnRows(rows)

			recs, err := db.GetReceptionsByPVZ(ctx, pvzId, nil, nil)
			assert.NoError(t, err)
			assert.Len(t, recs, 1)
			assert.Equal(t, receptionID, recs[0].ID)
			assert.Equal(t, recDate, recs[0].DateTime)
			assert.Equal(t, pvzId, recs[0].PVZId)
			assert.Equal(t, status, recs[0].Status)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("With Dates", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			db := NewPGXDatabase(mockPool)
			startDate, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
			endDate, _ := time.Parse(time.RFC3339, "2023-01-02T00:00:00Z")
			receptionID := uuid.New()
			recDate := time.Now().Add(-90 * time.Minute)
			status := "in_progress"

			expectedQuery := "SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id=$1 AND date_time >= $2 AND date_time <= $3 ORDER BY date_time DESC"
			mockPool.
				ExpectQuery(regexp.QuoteMeta(expectedQuery)).
				WithArgs(pvzId, startDate, endDate).
				WillReturnRows(pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID.String(), recDate, pvzId.String(), status))

			recs, err := db.GetReceptionsByPVZ(ctx, pvzId, &startDate, &endDate)
			assert.NoError(t, err)
			assert.Len(t, recs, 1)
			assert.Equal(t, receptionID, recs[0].ID)
			assert.Equal(t, recDate, recs[0].DateTime)
			assert.Equal(t, pvzId, recs[0].PVZId)
			assert.Equal(t, status, recs[0].Status)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	})
}

func TestGetProductsByReception(t *testing.T) {
	ctx := context.Background()
	receptionID := uuid.New()

	t.Run("GetProductsByReception", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		db := NewPGXDatabase(mockPool)
		product1ID := uuid.New()
		prod1Date := time.Now().Add(-30 * time.Minute)
		prod1Type := "sword"

		product2ID := uuid.New()
		prod2Date := time.Now().Add(-10 * time.Minute)
		prod2Type := "shield"

		rows := pgxmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
			AddRow(product1ID.String(), prod1Date, prod1Type, receptionID.String()).
			AddRow(product2ID.String(), prod2Date, prod2Type, receptionID.String())

		mockPool.
			ExpectQuery(regexp.QuoteMeta("SELECT id, date_time, type, reception_id FROM products WHERE reception_id=$1 ORDER BY date_time ASC")).
			WithArgs(receptionID).
			WillReturnRows(rows)

		products, err := db.GetProductsByReception(ctx, receptionID)
		assert.NoError(t, err)
		assert.Len(t, products, 2)

		assert.Equal(t, product1ID, products[0].ID)
		assert.Equal(t, prod1Date, products[0].DateTime)
		assert.Equal(t, prod1Type, products[0].Type)

		assert.Equal(t, product2ID, products[1].ID)
		assert.Equal(t, prod2Date, products[1].DateTime)
		assert.Equal(t, prod2Type, products[1].Type)

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestGetPVZ(t *testing.T) {
	ctx := context.Background()

	t.Run("GetPVZ", func(t *testing.T) {
		mockPool, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mockPool.Close()

		db := NewPGXDatabase(mockPool)

		id1 := "1"
		regDate1 := time.Now().Add(-1 * time.Hour)
		city1 := "Москва"

		id2 := "2"
		regDate2 := time.Now().Add(-2 * time.Hour)
		city2 := "Казань"

		rows := pgxmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow(id1, regDate1, city1).
			AddRow(id2, regDate2, city2)

		mockPool.
			ExpectQuery(regexp.QuoteMeta("SELECT id, registration_date, city FROM pvz")).
			WillReturnRows(rows)

		pvzs, err := db.GetPVZ(ctx)
		assert.NoError(t, err)
		assert.Len(t, pvzs, 2)

		assert.Equal(t, id1, pvzs[0].Id)
		assert.True(t, regDate1.Equal(pvzs[0].RegistrationDate.AsTime()))
		assert.Equal(t, city1, pvzs[0].City)

		assert.Equal(t, id2, pvzs[1].Id)
		assert.True(t, regDate2.Equal(pvzs[1].RegistrationDate.AsTime()))
		assert.Equal(t, city2, pvzs[1].City)

		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

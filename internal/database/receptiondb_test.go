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
)

func TestCloseLastReception(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()

	t.Run("CloseLastReception", func(t *testing.T) {
		t.Run("Begin error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			expectedErr := errors.New("begin error")
			mockPool.ExpectBegin().WillReturnError(expectedErr)

			db := NewPGXDatabase(mockPool)
			rec, err := db.CloseLastReception(ctx, pvzId)
			assert.Nil(t, rec)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Select query error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			expectedErr := errors.New("select error")
			mockPool.ExpectBegin()
			mockPool.
				ExpectQuery(regexp.QuoteMeta(`
		SELECT id, date_time, pvz_id, status
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
				WithArgs(pvzId).
				WillReturnError(expectedErr)
			mockPool.ExpectRollback()

			db := NewPGXDatabase(mockPool)
			rec, err := db.CloseLastReception(ctx, pvzId)
			assert.NotNil(t, rec)
			assert.Equal(t, uuid.Nil, rec.ID)
			assert.Equal(t, "", rec.Status)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Update query error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			receptionID := uuid.New()
			mockPool.ExpectBegin()
			rowsSelect := pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
				AddRow(receptionID.String(), time.Now(), pvzId.String(), "in_progress")
			mockPool.
				ExpectQuery(regexp.QuoteMeta(`
		SELECT id, date_time, pvz_id, status
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
				WithArgs(pvzId).
				WillReturnRows(rowsSelect)

			expectedErr := errors.New("update error")
			mockPool.
				ExpectQuery(regexp.QuoteMeta(`
		UPDATE receptions
		SET status='close'
		WHERE id=$1
		RETURNING id, date_time, pvz_id, status`)).
				WithArgs(receptionID).
				WillReturnError(expectedErr)
			mockPool.ExpectRollback()

			db := NewPGXDatabase(mockPool)
			rec, err := db.CloseLastReception(ctx, pvzId)
			assert.NotNil(t, rec)
			assert.Equal(t, receptionID, rec.ID)
			assert.Equal(t, "in_progress", rec.Status)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Commit error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			receptionID := uuid.New()
			mockPool.ExpectBegin()

			rowsSelect := pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
				AddRow(receptionID.String(), time.Now(), pvzId.String(), "in_progress")
			mockPool.
				ExpectQuery(regexp.QuoteMeta(`
		SELECT id, date_time, pvz_id, status
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
				WithArgs(pvzId).
				WillReturnRows(rowsSelect)

			rowsUpdate := pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
				AddRow(receptionID.String(), time.Now(), pvzId.String(), "close")
			mockPool.
				ExpectQuery(regexp.QuoteMeta(`
		UPDATE receptions
		SET status='close'
		WHERE id=$1
		RETURNING id, date_time, pvz_id, status`)).
				WithArgs(receptionID).
				WillReturnRows(rowsUpdate)
			expectedErr := errors.New("commit error")
			mockPool.ExpectCommit().WillReturnError(expectedErr)

			db := NewPGXDatabase(mockPool)
			rec, err := db.CloseLastReception(ctx, pvzId)

			assert.NotNil(t, rec)
			assert.Equal(t, "close", rec.Status)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Success", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			receptionID := uuid.New()
			mockPool.ExpectBegin()

			now := time.Now()
			rowsSelect := pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
				AddRow(receptionID.String(), now, pvzId.String(), "in_progress")
			mockPool.
				ExpectQuery(regexp.QuoteMeta(`
		SELECT id, date_time, pvz_id, status
		FROM receptions
		WHERE pvz_id=$1 AND status='in_progress'
		ORDER BY date_time DESC
		LIMIT 1`)).
				WithArgs(pvzId).
				WillReturnRows(rowsSelect)

			updatedTime := now.Add(5 * time.Minute)
			rowsUpdate := pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
				AddRow(receptionID.String(), updatedTime, pvzId.String(), "close")
			mockPool.
				ExpectQuery(regexp.QuoteMeta(`
		UPDATE receptions
		SET status='close'
		WHERE id=$1
		RETURNING id, date_time, pvz_id, status`)).
				WithArgs(receptionID).
				WillReturnRows(rowsUpdate)
			mockPool.ExpectCommit()

			db := NewPGXDatabase(mockPool)
			rec, err := db.CloseLastReception(ctx, pvzId)
			assert.NoError(t, err)
			assert.NotNil(t, rec)

			assert.Equal(t, "close", rec.Status)

			assert.True(t, updatedTime.Equal(rec.DateTime))
			assert.Equal(t, pvzId, rec.PVZId)
			assert.Equal(t, receptionID, rec.ID)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	})
}

func TestCreateReception(t *testing.T) {
	ctx := context.Background()
	pvzId := uuid.New()

	t.Run("CreateReception", func(t *testing.T) {
		t.Run("Begin error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			expectedErr := errors.New("begin error")
			mockPool.ExpectBegin().WillReturnError(expectedErr)

			db := NewPGXDatabase(mockPool)
			rec, err := db.CreateReception(ctx, pvzId)
			assert.Nil(t, rec)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Count query error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			mockPool.ExpectBegin()
			expectedErr := errors.New("count query error")
			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM receptions WHERE pvz_id=$1 AND status='in_progress'")).
				WithArgs(pvzId).
				WillReturnError(expectedErr)
			mockPool.ExpectRollback()

			db := NewPGXDatabase(mockPool)
			rec, err := db.CreateReception(ctx, pvzId)
			assert.Nil(t, rec)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Active reception exists", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			mockPool.ExpectBegin()

			rows := pgxmock.NewRows([]string{"count"}).AddRow(1)
			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM receptions WHERE pvz_id=$1 AND status='in_progress'")).
				WithArgs(pvzId).
				WillReturnRows(rows)
			mockPool.ExpectRollback()

			db := NewPGXDatabase(mockPool)
			rec, err := db.CreateReception(ctx, pvzId)
			assert.Nil(t, rec)
			assert.EqualError(t, err, "Активная приёмка уже существует")
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Insert error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			mockPool.ExpectBegin()

			rowsCount := pgxmock.NewRows([]string{"count"}).AddRow(0)
			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM receptions WHERE pvz_id=$1 AND status='in_progress'")).
				WithArgs(pvzId).
				WillReturnRows(rowsCount)
			expectedErr := errors.New("insert error")

			mockPool.
				ExpectQuery(regexp.QuoteMeta("INSERT INTO receptions (date_time, pvz_id, status) VALUES ($1, $2, $3) RETURNING id")).
				WithArgs(pgxmock.AnyArg(), pvzId, "in_progress").
				WillReturnError(expectedErr)
			mockPool.ExpectRollback()

			db := NewPGXDatabase(mockPool)
			rec, err := db.CreateReception(ctx, pvzId)
			assert.NotNil(t, rec)
			assert.Equal(t, "in_progress", rec.Status)
			assert.Equal(t, pvzId, rec.PVZId)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Commit error", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			mockPool.ExpectBegin()

			rowsCount := pgxmock.NewRows([]string{"count"}).AddRow(0)
			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM receptions WHERE pvz_id=$1 AND status='in_progress'")).
				WithArgs(pvzId).
				WillReturnRows(rowsCount)

			newReceptionID := uuid.New()
			rowsInsert := pgxmock.NewRows([]string{"id"}).AddRow(newReceptionID.String())
			mockPool.
				ExpectQuery(regexp.QuoteMeta("INSERT INTO receptions (date_time, pvz_id, status) VALUES ($1, $2, $3) RETURNING id")).
				WithArgs(pgxmock.AnyArg(), pvzId, "in_progress").
				WillReturnRows(rowsInsert)
			expectedErr := errors.New("commit error")
			mockPool.ExpectCommit().WillReturnError(expectedErr)

			db := NewPGXDatabase(mockPool)
			rec, err := db.CreateReception(ctx, pvzId)
			assert.NotNil(t, rec)
			assert.Equal(t, newReceptionID, rec.ID)
			assert.Equal(t, "in_progress", rec.Status)
			assert.EqualError(t, err, expectedErr.Error())
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})

		t.Run("Success", func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mockPool.Close()

			mockPool.ExpectBegin()

			rowsCount := pgxmock.NewRows([]string{"count"}).AddRow(0)
			mockPool.
				ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM receptions WHERE pvz_id=$1 AND status='in_progress'")).
				WithArgs(pvzId).
				WillReturnRows(rowsCount)

			newReceptionID := uuid.New()

			rowsInsert := pgxmock.NewRows([]string{"id"}).AddRow(newReceptionID.String())

			mockPool.
				ExpectQuery(regexp.QuoteMeta("INSERT INTO receptions (date_time, pvz_id, status) VALUES ($1, $2, $3) RETURNING id")).
				WithArgs(pgxmock.AnyArg(), pvzId, "in_progress").
				WillReturnRows(rowsInsert)
			mockPool.ExpectCommit()

			db := NewPGXDatabase(mockPool)
			rec, err := db.CreateReception(ctx, pvzId)
			assert.NoError(t, err)
			assert.NotNil(t, rec)

			assert.Equal(t, newReceptionID, rec.ID)
			assert.Equal(t, "in_progress", rec.Status)
			assert.Equal(t, pvzId, rec.PVZId)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	})
}

package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"

	"pvz/internal/models"
)

func TestCreateUser(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Не удалось создать pgxmock pool: %v", err)
	}
	defer mockPool.Close()

	db := NewPGXDatabase(mockPool)

	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "employee",
	}

	expectedID := uuid.New()

	mockPool.
		ExpectQuery("INSERT INTO users \\(email, password, role\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id").
		WithArgs(user.Email, user.Password, user.Role).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedID.String()))

	err = db.CreateUser(context.Background(), user)
	assert.NoError(t, err, "Ожидалась успешная вставка пользователя")
	assert.Equal(t, expectedID, user.ID, "Полученный ID не совпадает с ожидаемым")

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %s", err)
	}
}

func TestGetUserByEmail(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Не удалось создать pgxmock pool: %v", err)
	}
	defer mockPool.Close()

	db := NewPGXDatabase(mockPool)

	email := "test@example.com"
	expectedID := uuid.New()
	expectedPassword := "hashedpassword"
	expectedRole := "employee"

	mockPool.
		ExpectQuery("SELECT id, email, password, role FROM users WHERE email=\\$1").
		WithArgs(email).
		WillReturnRows(
			pgxmock.NewRows([]string{"id", "email", "password", "role"}).
				AddRow(expectedID.String(), email, expectedPassword, expectedRole),
		)

	user, err := db.GetUserByEmail(context.Background(), email)
	assert.NoError(t, err, "Ожидалась успешная выборка пользователя")
	assert.NotNil(t, user, "Пользователь не должен быть nil")
	assert.Equal(t, expectedID, user.ID, "ID пользователя не совпадает с ожидаемым")
	assert.Equal(t, email, user.Email, "Email пользователя не совпадает")
	assert.Equal(t, expectedPassword, user.Password, "Пароль пользователя не совпадает")
	assert.Equal(t, expectedRole, user.Role, "Роль пользователя не совпадает")

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %s", err)
	}
}

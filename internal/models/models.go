package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Email    string    `json:"email" db:"email"`
	Password string    `json:"-"`
	Role     string    `json:"role" db:"role"`
}

type PVZ struct {
	ID               uuid.UUID `json:"id" db:"id"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date"`
	City             string    `json:"city" db:"city"`
}

type Reception struct {
	ID       uuid.UUID `json:"id" db:"id"`
	DateTime time.Time `json:"dateTime" db:"date_time"`
	PVZId    uuid.UUID `json:"pvzId" db:"pvz_id"`
	Status   string    `json:"status" db:"status"`
}

type Product struct {
	ID          uuid.UUID `json:"id" db:"id"`
	DateTime    time.Time `json:"dateTime" db:"date_time"`
	Type        string    `json:"type" db:"type"`
	ReceptionId uuid.UUID `json:"receptionId" db:"reception_id"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type PVZResponse struct {
	PVZ        *PVZ             `json:"pvz"`
	Receptions []*ReceptionInfo `json:"receptions"`
}

type ReceptionInfo struct {
	Reception *Reception `json:"reception"`
	Products  []*Product `json:"products"`
}

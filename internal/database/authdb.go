package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"pvz/internal/models"
	pb "pvz/internal/pb/pvz_v1"
)

type Database interface {
	CreateUser(ctx context.Context, user *models.User) (err error)
	GetUserByEmail(ctx context.Context, email string) (user *models.User, err error)
	CreatePVZ(ctx context.Context, pvz *models.PVZ) (err error)
	GetPVZs(ctx context.Context, limit, offset int) (pvzs []models.PVZ, err error)
	GetReceptionsByPVZ(ctx context.Context, pvzId uuid.UUID, startDate, endDate *time.Time) (recs []models.Reception, err error)
	GetProductsByReception(ctx context.Context, receptionID uuid.UUID) (products []*models.Product, err error)
	CloseLastReception(ctx context.Context, pvzId uuid.UUID) (rec *models.Reception, err error)
	DeleteLastProduct(ctx context.Context, pvzId uuid.UUID) (err error)
	CreateReception(ctx context.Context, pvzId uuid.UUID) (rec *models.Reception, err error)
	AddProduct(ctx context.Context, pvzId uuid.UUID, productType string) (product *models.Product, err error)
	GetPVZ(ctx context.Context) (pvzs []*pb.PVZ, err error)
}

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PGXDatabase struct {
	pool DBPool
}

func NewPGXDatabase(pool DBPool) *PGXDatabase {
	return &PGXDatabase{pool: pool}
}

func (db *PGXDatabase) CreateUser(ctx context.Context, user *models.User) (err error) {
	query := `INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`
	return db.pool.QueryRow(ctx, query, user.Email, user.Password, user.Role).Scan(&user.ID)
}

func (db *PGXDatabase) GetUserByEmail(ctx context.Context, email string) (user *models.User, err error) {
	user = &models.User{}
	query := `SELECT id, email, password, role FROM users WHERE email=$1`
	err = db.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	return user, err
}

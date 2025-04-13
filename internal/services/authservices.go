package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"pvz/internal/database"
	"pvz/internal/models"
	pb "pvz/internal/pb/pvz_v1"
)

type ServiceInterface interface {
	DummyLogin(req *models.DummyLoginRequest) (token string, status int, err error)
	Register(ctx context.Context, req *models.RegisterRequest) (ans *models.User, status int, err error)
	Login(ctx context.Context, req *models.LoginRequest) (token string, status int, err error)
	CreatePVZ(ctx context.Context, pvz *models.PVZ, role string) (int, error)
	ListPVZ(ctx context.Context, startDateStr string, endDateStr string, pageStr string, limitStr string) (results []*models.PVZResponse, status int, err error)
	CloseLastReception(ctx context.Context, role string, pvzId uuid.UUID) (rec *models.Reception, status int, err error)
	DeleteLastProduct(ctx context.Context, role string, pvzId uuid.UUID) (status int, err error)
	CreateReception(ctx context.Context, role string, pvzId uuid.UUID) (rec *models.Reception, status int, err error)
	AddProduct(ctx context.Context, role string, pvzId uuid.UUID, producttype string) (product *models.Product, status int, err error)
	GetPVZ(ctx context.Context) (pvzs []*pb.PVZ, err error)
}

type Service struct {
	database  database.Database
	jwtSecret []byte
}

func NewService(db database.Database, jwtSecret []byte) *Service {
	return &Service{database: db, jwtSecret: jwtSecret}
}

func (s *Service) generateToken(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"id":   userID.String(),
		"role": role,
		"exp":  time.Now().Add(72 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) DummyLogin(req *models.DummyLoginRequest) (token string, status int, err error) {
	if req.Role != "employee" && req.Role != "moderator" {
		return "", http.StatusBadRequest, errors.New("неверная роль")
	}
	userID := uuid.New()
	token, err = s.generateToken(userID, req.Role)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("ошибка генерации токена")
	}
	return token, http.StatusOK, nil
}

func (s *Service) Register(ctx context.Context, req *models.RegisterRequest) (ans *models.User, status int, err error) {
	if req.Role != "employee" && req.Role != "moderator" {
		return ans, http.StatusBadRequest, errors.New("неверная роль")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ans, http.StatusInternalServerError, errors.New("ошибка хэширования пароля")
	}
	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
	}
	if err := s.database.CreateUser(ctx, &user); err != nil {
		return ans, http.StatusBadRequest, errors.New(fmt.Sprintf("ошибка регистрации: %v", err))
	}
	ans = &models.User{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}
	return ans, http.StatusOK, nil
}

func (s *Service) Login(ctx context.Context, req *models.LoginRequest) (token string, status int, err error) {
	user, err := s.database.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", http.StatusUnauthorized, errors.New("неверные учетные данные")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", http.StatusUnauthorized, errors.New("неверные учетные данные")
	}
	token, err = s.generateToken(user.ID, user.Role)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("ошибка генерации токена")
	}
	return token, http.StatusOK, nil
}

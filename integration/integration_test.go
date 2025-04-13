package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"

	"pvz/internal/database"
	"pvz/internal/middleware"
	"pvz/internal/models"
	"pvz/internal/services"
	"pvz/internal/transport/rest"
)

var testServerURL string
var postgresInstance *embeddedpostgres.EmbeddedPostgres
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	cfg := embeddedpostgres.DefaultConfig().
		Port(65530).
		Version(embeddedpostgres.V13)

	postgresInstance = embeddedpostgres.NewDatabase(cfg)
	if err := postgresInstance.Start(); err != nil {
		log.Fatalf("Не удалось запустить embedded postgres: %v", err)
	}

	connStr := cfg.GetConnectionURL()
	var err error
	testPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось создать pgxpool: %v", err)
	}

	time.Sleep(time.Second)

	if err := runMigrations("../migrations/init.sql"); err != nil {
		log.Fatalf("Ошибка при миграциях: %v", err)
	}

	db := database.NewPGXDatabase(testPool)
	jwtSecret := []byte("testsecret")
	svc := services.NewService(db, jwtSecret)
	mw := middleware.NewMiddleware(jwtSecret)
	h := rest.NewHandler(svc)

	r := mux.NewRouter()

	r.HandleFunc("/dummyLogin", h.DummyLoginHandler).Methods("POST")
	r.HandleFunc("/register", h.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", h.LoginHandler).Methods("POST")

	r.Handle("/pvz", mw.AuthMiddleware(http.HandlerFunc(h.CreatePVZHandler))).Methods("POST")
	r.Handle("/pvz", mw.AuthMiddleware(http.HandlerFunc(h.ListPVZHandler))).Methods("GET")
	r.Handle("/pvz/{pvzId}/close_last_reception", mw.AuthMiddleware(http.HandlerFunc(h.CloseLastReceptionHandler))).Methods("POST")
	r.Handle("/pvz/{pvzId}/delete_last_product", mw.AuthMiddleware(http.HandlerFunc(h.DeleteLastProductHandler))).Methods("POST")

	r.Handle("/receptions", mw.AuthMiddleware(http.HandlerFunc(h.CreateReceptionHandler))).Methods("POST")
	r.Handle("/products", mw.AuthMiddleware(http.HandlerFunc(h.AddProductHandler))).Methods("POST")

	ts := httptest.NewServer(r)
	testServerURL = ts.URL

	code := m.Run()

	ts.Close()
	testPool.Close()
	_ = postgresInstance.Stop()

	os.Exit(code)
}

func runMigrations(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = testPool.Exec(context.Background(), string(data))
	return err
}

func TestHandlersFullFlow(t *testing.T) {
	var modToken, empToken string
	var pvzId uuid.UUID

	t.Run("Модератор DummyLogin", func(t *testing.T) {
		body := []byte(`{"role":"moderator"}`)
		resp, err := http.Post(testServerURL+"/dummyLogin", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var token string
		json.NewDecoder(resp.Body).Decode(&token)
		assert.NotEmpty(t, token)
		modToken = token
	})

	t.Run("Создать ПВЗ (модератор)", func(t *testing.T) {
		body := []byte(`{"city":"Москва"}`)
		req, _ := http.NewRequest(http.MethodPost, testServerURL+"/pvz", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+modToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var created models.PVZ
		json.NewDecoder(resp.Body).Decode(&created)
		assert.NotEqual(t, uuid.Nil, created.ID)
		pvzId = created.ID
		t.Logf("Создан ПВЗ: %s", pvzId.String())
	})

	t.Run("Сотрудник DummyLogin", func(t *testing.T) {
		body := []byte(`{"role":"employee"}`)
		resp, err := http.Post(testServerURL+"/dummyLogin", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var token string
		json.NewDecoder(resp.Body).Decode(&token)
		assert.NotEmpty(t, token)
		empToken = token
	})

	var receptionId uuid.UUID
	t.Run("Создать приемку", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{"pvzId":"%s"}`, pvzId.String()))
		req, _ := http.NewRequest(http.MethodPost, testServerURL+"/receptions", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+empToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var rec models.Reception
		json.NewDecoder(resp.Body).Decode(&rec)
		assert.NotEqual(t, uuid.Nil, rec.ID)
		receptionId = rec.ID
		t.Logf("Создана приемка: %s", receptionId.String())
	})

	t.Run("Добавить 50 товаров в приемку", func(t *testing.T) {
		client := &http.Client{}

		for i := 1; i <= 50; i++ {
			body := []byte(fmt.Sprintf(`{"type":"электроника","pvzId":"%s"}`, pvzId.String()))
			req, _ := http.NewRequest(http.MethodPost, testServerURL+"/products", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+empToken)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var prod models.Product
			json.NewDecoder(resp.Body).Decode(&prod)
			assert.NotEqual(t, uuid.Nil, prod.ID)
			t.Logf("Добавлен товар №%d: %s", i, prod.ID.String())
		}
	})

	t.Run("Закрыть приемку", func(t *testing.T) {
		client := &http.Client{}
		url := fmt.Sprintf("%s/pvz/%s/close_last_reception", testServerURL, pvzId.String())
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Authorization", "Bearer "+empToken)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var closedRec models.Reception
		json.NewDecoder(resp.Body).Decode(&closedRec)
		assert.Equal(t, "close", closedRec.Status)
		t.Logf("Приемка %s закрыта. Статус: %s", closedRec.ID, closedRec.Status)
	})
}

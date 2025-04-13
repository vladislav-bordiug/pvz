package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"

	"pvz/internal/contextkeys"
)

func TestAuthMiddleware(t *testing.T) {
	secret := []byte("testsecret")
	mw := NewMiddleware(secret)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(contextkeys.ContextKeyUserID)
		role := r.Context().Value(contextkeys.ContextKeyRole)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("user:" + userID.(string) + ", role:" + role.(string)))
	})

	handlerToTest := mw.AuthMiddleware(nextHandler)

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Отсутствует заголовок Authorization")
	})

	t.Run("Invalid Header Format", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic token")
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Неверный формат заголовка")
	})

	t.Run("Invalid Token", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.value")
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Неверный токен")
	})

	t.Run("Missing id Claim", func(t *testing.T) {

		claims := jwt.MapClaims{
			"role": "employee",
			"exp":  time.Now().Add(time.Hour).Unix(),
			"iat":  time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secret)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Неверный id")
	})

	t.Run("Missing role Claim", func(t *testing.T) {

		claims := jwt.MapClaims{
			"id":  "user123",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secret)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Неверная роль")
	})

	t.Run("Success", func(t *testing.T) {

		claims := jwt.MapClaims{
			"id":   "user123",
			"role": "employee",
			"exp":  time.Now().Add(time.Hour).Unix(),
			"iat":  time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secret)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		response := rr.Body.String()
		assert.Equal(t, http.StatusOK, rr.Code)

		assert.True(t, strings.Contains(response, "user:user123"))
		assert.True(t, strings.Contains(response, "role:employee"))
	})
}

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"

	"pvz/internal/contextkeys"
)

type Middleware struct {
	jwtSecret []byte
}

func NewMiddleware(jwtSecret []byte) *Middleware {
	return &Middleware{jwtSecret: jwtSecret}
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Отсутствует заголовок Authorization", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Неверный формат заголовка", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
			}
			return m.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Неверные claims", http.StatusUnauthorized)
			return
		}
		idStr, ok := claims["id"].(string)
		if !ok {
			http.Error(w, "Неверный id", http.StatusUnauthorized)
			return
		}
		role, ok := claims["role"].(string)
		if !ok {
			http.Error(w, "Неверная роль", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextkeys.ContextKeyUserID, idStr)
		ctx = context.WithValue(ctx, contextkeys.ContextKeyRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

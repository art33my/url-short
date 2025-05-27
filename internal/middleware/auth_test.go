package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"url-short/internal/config"
	"url-short/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test_secret"}

	t.Run("Valid token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": 42,
			"exp":     time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", tokenString)

		middlewareFunc := middleware.AuthMiddleware(cfg)
		middlewareFunc(c)

		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, 42, userID)
		assert.False(t, c.IsAborted())
	})

	t.Run("Missing token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		middlewareFunc := middleware.AuthMiddleware(cfg)
		middlewareFunc(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("Invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "invalid_token")

		middlewareFunc := middleware.AuthMiddleware(cfg)
		middlewareFunc(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})
}

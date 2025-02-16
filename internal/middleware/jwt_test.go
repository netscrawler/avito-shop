package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthMiddleware(t *testing.T) {
	// Отключаем режим Gin по умолчанию для тестов
	gin.SetMode(gin.TestMode)

	const testSecret = "test-secret"

	t.Run("успешная аутентификация", func(t *testing.T) {
		// Создаем валидный токен
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "testuser",
			"exp":      time.Now().Add(time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte(testSecret))
		assert.NoError(t, err)

		// Создаем тестовый роутер
		r := gin.New()
		r.Use(JWTAuthMiddleware(testSecret))
		r.GET("/test", func(c *gin.Context) {
			username := c.GetString("username")
			c.JSON(http.StatusOK, gin.H{"username": username})
		})

		// Создаем тестовый запрос
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "testuser")
	})

	t.Run("отсутствие токена", func(t *testing.T) {
		r := gin.New()
		r.Use(JWTAuthMiddleware(testSecret))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("некорректный формат токена", func(t *testing.T) {
		r := gin.New()
		r.Use(JWTAuthMiddleware(testSecret))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Invalid-Token-Format")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("истекший токен", func(t *testing.T) {
		// Создаем токен с истекшим временем
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "testuser",
			"exp":      time.Now().Add(-time.Hour).Unix(), // Токен истек час назад
		})
		tokenString, err := token.SignedString([]byte(testSecret))
		assert.NoError(t, err)

		r := gin.New()
		r.Use(JWTAuthMiddleware(testSecret))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("некорректная подпись токена", func(t *testing.T) {
		// Создаем токен с другим секретом
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "testuser",
			"exp":      time.Now().Add(time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte("wrong-secret"))
		assert.NoError(t, err)

		r := gin.New()
		r.Use(JWTAuthMiddleware(testSecret))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

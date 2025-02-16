package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	// Отключаем режим Gin по умолчанию для тестов
	gin.SetMode(gin.TestMode)

	t.Run("логирование успешного запроса", func(t *testing.T) {
		// Создаем буфер для логов
		var logBuffer bytes.Buffer
		logger := logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetOutput(&logBuffer)

		// Создаем тестовый роутер
		r := gin.New()
		r.Use(LoggerMiddleware(logger))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		})

		// Создаем тестовый запрос
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		r.ServeHTTP(w, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем наличие логов
		logEntries := strings.Split(strings.TrimSpace(logBuffer.String()), "\n")
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(logEntries[0]), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "Request processed", logEntry["msg"])
		assert.Equal(t, float64(200), logEntry["status"])
		assert.Equal(t, "GET", logEntry["method"])
		assert.Equal(t, "/test", logEntry["path"])
	})

	t.Run("логирование запроса с телом", func(t *testing.T) {
		var logBuffer bytes.Buffer
		logger := logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetOutput(&logBuffer)

		r := gin.New()
		r.Use(LoggerMiddleware(logger))
		r.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"status": "created"})
		})

		body := bytes.NewBufferString(`{"test": "data"}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		logEntries := strings.Split(strings.TrimSpace(logBuffer.String()), "\n")
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(logEntries[0]), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "Request processed", logEntry["msg"])
		assert.Equal(t, float64(201), logEntry["status"])

		request, ok := logEntry["request"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "data", request["test"])
	})

	t.Run("логирование ошибки клиента", func(t *testing.T) {
		var logBuffer bytes.Buffer
		logger := logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetOutput(&logBuffer)

		r := gin.New()
		r.Use(LoggerMiddleware(logger))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		logEntries := strings.Split(strings.TrimSpace(logBuffer.String()), "\n")
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(logEntries[0]), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "Client error", logEntry["msg"])
		assert.Equal(t, float64(400), logEntry["status"])
	})

	t.Run("логирование ошибки сервера", func(t *testing.T) {
		var logBuffer bytes.Buffer
		logger := logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetOutput(&logBuffer)

		r := gin.New()
		r.Use(LoggerMiddleware(logger))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		logEntries := strings.Split(strings.TrimSpace(logBuffer.String()), "\n")
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(logEntries[0]), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, "Server error", logEntry["msg"])
		assert.Equal(t, float64(500), logEntry["status"])
	})
}

func TestResponseBodyWriter(t *testing.T) {
	t.Run("запись тела ответа", func(t *testing.T) {
		originalBody := []byte("test response")
		buffer := &bytes.Buffer{}
		writer := &responseBodyWriter{
			ResponseWriter: &mockResponseWriter{},
			body:           buffer,
		}

		n, err := writer.Write(originalBody)

		assert.NoError(t, err)
		assert.Equal(t, len(originalBody), n)
		assert.Equal(t, originalBody, buffer.Bytes())
	})
}

// Мок для ResponseWriter
type mockResponseWriter struct {
	gin.ResponseWriter
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

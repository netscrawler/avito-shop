package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	TraceIDKey = "trace_id"
	MaskString = "********"
)

// Список полей, которые нужно маскировать
var sensitiveFields = []string{
	"password",
	"token",
	"access_token",
	"refresh_token",
	"passport",
	"card_number",
	"cvv",
	"secret",
}

// responseBodyWriter реализует интерфейс gin.ResponseWriter и сохраняет тело ответа
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// maskSensitiveData маскирует конфиденциальные данные в JSON
func maskSensitiveData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			lowercaseKey := strings.ToLower(key)
			shouldMask := false
			for _, sensitive := range sensitiveFields {
				if strings.Contains(lowercaseKey, sensitive) {
					shouldMask = true
					break
				}
			}
			if shouldMask {
				result[key] = MaskString
			} else {
				result[key] = maskSensitiveData(value)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = maskSensitiveData(value)
		}
		return result
	default:
		return v
	}
}

// LoggerMiddleware создает middleware для логирования запросов и ответов
func LoggerMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Генерируем trace_id для запроса
		traceID := uuid.New().String()
		c.Set(TraceIDKey, traceID)

		start := time.Now()

		// Читаем тело запроса
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Создаем буфер для тела ответа
		responseBody := &bytes.Buffer{}
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = writer

		// Продолжаем обработку запроса
		c.Next()

		// Форматируем тела запроса и ответа для логирования
		var requestJSON, responseJSON interface{}
		if len(requestBody) > 0 {
			if err := json.Unmarshal(requestBody, &requestJSON); err != nil {
				requestJSON = string(requestBody)
			} else {
				requestJSON = maskSensitiveData(requestJSON)
			}
		}
		if responseBody.Len() > 0 {
			if err := json.Unmarshal(responseBody.Bytes(), &responseJSON); err != nil {
				responseJSON = responseBody.String()
			} else {
				responseJSON = maskSensitiveData(responseJSON)
			}
		}

		// Создаем запись в логе
		fields := logrus.Fields{
			"timestamp":  time.Now().Format(time.RFC3339),
			"trace_id":   traceID,
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"latency_ms": time.Since(start).Milliseconds(),
			"user_agent": c.Request.UserAgent(),
		}

		// Добавляем тела запроса и ответа, если они есть
		if requestJSON != nil {
			fields["request"] = requestJSON
		}
		if responseJSON != nil {
			fields["response"] = responseJSON
		}

		// Добавляем информацию о пользователе, если есть
		if username, exists := c.Get("username"); exists {
			fields["username"] = username
		}

		// Логируем с соответствующим уровнем в зависимости от статуса
		entry := log.WithFields(fields)
		switch {
		case c.Writer.Status() >= 500:
			entry.Error("Server error")
		case c.Writer.Status() >= 400:
			entry.Warn("Client error")
		default:
			entry.Info("Request processed")
		}
	}
}

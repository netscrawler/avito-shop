package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("загрузка конфигурации по умолчанию", func(t *testing.T) {
		cfg, err := New()
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// Проверяем значения по умолчанию
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, 5*time.Second, cfg.Server.ReadTimeout)
		assert.Equal(t, 5*time.Second, cfg.Server.WriteTimeout)
	})

	t.Run("загрузка конфигурации из переменных окружения", func(t *testing.T) {
		// Устанавливаем переменные окружения
		os.Setenv("SERVER_PORT", "3000")
		os.Setenv("JWT_SECRET", "test-secret")
		os.Setenv("DATABASE_HOST", "testhost")
		os.Setenv("DATABASE_PORT", "5432")
		os.Setenv("DATABASE_USER", "testuser")
		os.Setenv("DATABASE_PASSWORD", "testpass")
		os.Setenv("DATABASE_NAME", "testdb")
		defer func() {
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("JWT_SECRET")
			os.Unsetenv("DATABASE_HOST")
			os.Unsetenv("DATABASE_PORT")
			os.Unsetenv("DATABASE_USER")
			os.Unsetenv("DATABASE_PASSWORD")
			os.Unsetenv("DATABASE_NAME")
		}()

		cfg, err := New()
		require.NoError(t, err)
		assert.NotNil(t, cfg)

		// Проверяем значения из переменных окружения
		assert.Equal(t, "3000", cfg.Server.Port)
		assert.Equal(t, "test-secret", cfg.JWT.Secret)
		assert.Equal(t, "testhost", cfg.Database.Host)
		assert.Equal(t, "5432", cfg.Database.Port)
		assert.Equal(t, "testuser", cfg.Database.User)
		assert.Equal(t, "testpass", cfg.Database.Password)
		assert.Equal(t, "testdb", cfg.Database.DBName)
	})
}

func TestDatabaseConfig(t *testing.T) {
	t.Run("формирование URL базы данных", func(t *testing.T) {
		cfg := &Config{
			Database: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "password",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
		}

		expectedURL := "postgres://postgres:password@localhost:5432/testdb?sslmode=disable"
		assert.Equal(t, expectedURL, cfg.Database.DatabaseUrl())
	})

	t.Run("формирование URL с пустым паролем", func(t *testing.T) {
		cfg := &Config{
			Database: DatabaseConfig{
				Host:    "localhost",
				Port:    "5432",
				User:    "postgres",
				DBName:  "testdb",
				SSLMode: "disable",
			},
		}

		expectedURL := "postgres://postgres:@localhost:5432/testdb?sslmode=disable"
		assert.Equal(t, expectedURL, cfg.Database.DatabaseUrl())
	})
}

func TestServerConfig(t *testing.T) {
	t.Run("проверка таймаутов сервера", func(t *testing.T) {
		os.Setenv("SERVER_READ_TIMEOUT", "20s")
		os.Setenv("SERVER_WRITE_TIMEOUT", "15s")
		defer func() {
			os.Unsetenv("SERVER_READ_TIMEOUT")
			os.Unsetenv("SERVER_WRITE_TIMEOUT")
		}()

		cfg, err := New()
		require.NoError(t, err)

		assert.Equal(t, 20*time.Second, cfg.Server.ReadTimeout)
		assert.Equal(t, 15*time.Second, cfg.Server.WriteTimeout)
	})

	t.Run("некорректный формат таймаута", func(t *testing.T) {
		os.Setenv("SERVER_READ_TIMEOUT", "invalid")
		defer os.Unsetenv("SERVER_READ_TIMEOUT")

		cfg, err := New()
		require.NoError(t, err)
		// При некорректном формате должно использоваться значение по умолчанию
		assert.Equal(t, 5*time.Second, cfg.Server.ReadTimeout)
	})
}

func TestJWTConfig(t *testing.T) {
	t.Run("проверка секрета JWT", func(t *testing.T) {
		defaultSecret := "your-256-bit-secret"
		os.Setenv("JWT_SECRET", "")
		defer os.Unsetenv("JWT_SECRET")

		cfg, err := New()
		require.NoError(t, err)
		// Если JWT_SECRET пустой, должно использоваться значение по умолчанию
		assert.Equal(t, defaultSecret, cfg.JWT.Secret)
	})

	t.Run("проверка пользовательского секрета JWT", func(t *testing.T) {
		customSecret := "custom-secret"
		os.Setenv("JWT_SECRET", customSecret)
		defer os.Unsetenv("JWT_SECRET")

		cfg, err := New()
		require.NoError(t, err)
		assert.Equal(t, customSecret, cfg.JWT.Secret)
	})
}

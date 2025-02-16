package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/repository"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultInitialCoins = uint64(1000)
	defaultTokenTTL     = 72 * time.Hour
)

// UserService предоставляет методы для работы с пользователями
type userService struct {
	repo      repository.UserRepository
	jwtSecret string
}

// NewUserService создает новый экземпляр сервиса пользователей
func NewUserService(repo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// GetUserByUsername получает пользователя по имени
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	const op = "UserService.GetUserByUsername"

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// RegisterUser регистрирует нового пользователя
func (s *userService) RegisterUser(ctx context.Context, username, password string) error {
	const op = "UserService.RegisterUser"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: хеширование пароля: %w", op, err)
	}

	user := domain.NewUser(username, hashedPassword, defaultInitialCoins)
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("%s: создание пользователя: %w", op, err)
	}

	return nil
}

// AuthenticateUser аутентифицирует пользователя и возвращает JWT токен
func (s *userService) AuthenticateUser(ctx context.Context, username, password string) (string, error) {
	const op = "UserService.AuthenticateUser"

	// Пытаемся получить пользователя
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// Пытаемся зарегистрировать пользователя
			logrus.Infof("%s: пользователь не найден, выполняем регистрацию", op)
			if regErr := s.RegisterUser(ctx, username, password); regErr != nil {
				if regErr == domain.ErrUserAlreadyExists {
					// Если во время регистрации обнаружили, что пользователь уже существует,
					// значит он был создан параллельным запросом
					user, err = s.repo.GetUserByUsername(ctx, username)
					if err != nil {
						logrus.Errorf("%s: ошибка получения существующего пользователя: %v", op, err)
						return "", fmt.Errorf("%s: получение существующего пользователя: %w", op, err)
					}
				} else {
					logrus.Errorf("%s: ошибка регистрации пользователя: %v", op, regErr)
					return "", fmt.Errorf("%s: регистрация нового пользователя: %w", op, regErr)
				}
			} else {
				// Регистрация прошла успешно, получаем созданного пользователя
				user, err = s.repo.GetUserByUsername(ctx, username)
				if err != nil {
					logrus.Errorf("%s: ошибка получения пользователя после регистрации: %v", op, err)
					return "", fmt.Errorf("%s: получение пользователя после регистрации: %w", op, err)
				}
			}
		} else {
			logrus.Errorf("%s: ошибка получения пользователя: %v", op, err)
			return "", fmt.Errorf("%s: получение пользователя: %w", op, err)
		}
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		logrus.Errorf("%s: неверный пароль для пользователя %s", op, username)
		return "", domain.ErrInvalidCredentials
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		logrus.Errorf("%s: ошибка подписи токена: %v", op, err)
		return "", fmt.Errorf("%s: подпись токена: %w", op, err)
	}

	return tokenString, nil
}

// GetUserInfo возвращает информацию о пользователе
func (s *userService) GetUserInfo(ctx context.Context, username string) (*domain.User, error) {
	const op = "UserService.GetUserInfo"

	user, err := s.repo.GetUserInfo(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

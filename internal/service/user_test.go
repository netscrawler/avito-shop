package service

import (
	"context"
	"testing"

	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	service := NewUserService(userRepo, "test-secret")

	username := "testuser"
	password := "password123"

	userRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	// Act
	err := service.RegisterUser(ctx, username, password)

	// Assert
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthenticateUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	service := NewUserService(userRepo, "test-secret")

	username := "testuser"
	password := "password123"

	// Настраиваем мок для автоматической регистрации
	userRepo.On("GetUserByUsername", mock.Anything, username).Return(nil, domain.ErrUserNotFound).Once()
	userRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	// Настраиваем мок для возврата пользователя после регистрации
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userRepo.On("GetUserByUsername", mock.Anything, username).Return(&domain.User{
		Username: username,
		Password: hashedPassword,
	}, nil)

	// Act
	token, err := service.AuthenticateUser(ctx, username, password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	userRepo.AssertExpectations(t)
}

func TestAuthenticateUser_InvalidCredentials(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	service := NewUserService(userRepo, "test-secret")

	username := "testuser"
	wrongPassword := "wrongpassword"

	// Настраиваем мок для существующего пользователя
	user := &domain.User{
		Username: username,
		Password: []byte("$2a$10$..."), // Хеш пароля будет создан автоматически
	}
	userRepo.On("GetUserByUsername", mock.Anything, username).Return(user, nil)

	// Act
	token, err := service.AuthenticateUser(ctx, username, wrongPassword)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}

func TestGetUserInfo_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	service := NewUserService(userRepo, "test-secret")

	username := "testuser"
	expectedUser := &domain.User{
		Id:       1,
		Username: username,
		Coins:    1000,
		Inventory: []domain.UserInventory{
			{Type: "item1", Quantity: 1},
		},
	}

	userRepo.On("GetUserInfo", mock.Anything, username).Return(expectedUser, nil)

	// Act
	user, err := service.GetUserInfo(ctx, username)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	userRepo.AssertExpectations(t)
}

func TestGetUserInfo_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	service := NewUserService(userRepo, "test-secret")

	username := "nonexistent"

	userRepo.On("GetUserInfo", mock.Anything, username).Return(nil, domain.ErrUserNotFound)

	// Act
	user, err := service.GetUserInfo(ctx, username)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	userRepo.AssertExpectations(t)
}

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моки сервисов
type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) RegisterUser(ctx context.Context, username, password string) error {
	args := m.Called(ctx, username, password)
	return args.Error(0)
}

func (m *mockUserService) AuthenticateUser(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func (m *mockUserService) GetUserInfo(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*domain.User), args.Error(1)
}

type mockTransferService struct {
	mock.Mock
}

func (m *mockTransferService) SendCoins(ctx context.Context, sender, receiver string, amount uint64) error {
	args := m.Called(ctx, sender, receiver, amount)
	return args.Error(0)
}

func (m *mockTransferService) GetTransactionHistory(ctx context.Context, username string) (model.CoinHistory, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(model.CoinHistory), args.Error(1)
}

type mockMerchService struct {
	mock.Mock
}

func (m *mockMerchService) BuyMerch(ctx context.Context, username, merchName string) error {
	args := m.Called(ctx, username, merchName)
	return args.Error(0)
}

func (m *mockMerchService) GetAllMerch(ctx context.Context) ([]*domain.Merch, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Merch), args.Error(1)
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	return c, w
}

func TestHealthCheck(t *testing.T) {
	c, w := setupTestContext()
	h := NewHandler(&mockUserService{}, &mockTransferService{}, &mockMerchService{})

	h.HealthCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestAuthenticate(t *testing.T) {
	t.Run("успешная аутентификация", func(t *testing.T) {
		userService := new(mockUserService)
		h := NewHandler(userService, &mockTransferService{}, &mockMerchService{})

		userService.On("AuthenticateUser", mock.Anything, "testuser", "password").Return("test-token", nil)

		c, w := setupTestContext()
		body := bytes.NewBufferString(`{"username":"testuser","password":"password"}`)
		c.Request = httptest.NewRequest("POST", "/auth", body)
		c.Request.Header.Set("Content-Type", "application/json")

		h.Authenticate(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "test-token", response["token"])
		userService.AssertExpectations(t)
	})

	t.Run("неверные учетные данные", func(t *testing.T) {
		userService := new(mockUserService)
		h := NewHandler(userService, &mockTransferService{}, &mockMerchService{})

		userService.On("AuthenticateUser", mock.Anything, "testuser", "wrongpass").
			Return("", domain.ErrInvalidCredentials)

		c, w := setupTestContext()
		body := bytes.NewBufferString(`{"username":"testuser","password":"wrongpass"}`)
		c.Request = httptest.NewRequest("POST", "/auth", body)
		c.Request.Header.Set("Content-Type", "application/json")

		h.Authenticate(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		userService.AssertExpectations(t)
	})
}

func TestGetInfo(t *testing.T) {
	t.Run("успешное получение информации", func(t *testing.T) {
		userService := new(mockUserService)
		transferService := new(mockTransferService)
		h := NewHandler(userService, transferService, &mockMerchService{})

		user := &domain.User{
			Username: "testuser",
			Coins:    1000,
			Inventory: []domain.UserInventory{
				{Type: "item1", Quantity: 1},
			},
		}

		history := model.CoinHistory{
			Sent:     []model.SentTransaction{{ToUser: "user1", Amount: 100}},
			Received: []model.ReceivedTransaction{{FromUser: "user2", Amount: 200}},
		}

		userService.On("GetUserInfo", mock.Anything, "testuser").Return(user, nil)
		transferService.On("GetTransactionHistory", mock.Anything, "testuser").Return(history, nil)

		c, w := setupTestContext()
		c.Set("username", "testuser")
		c.Request = httptest.NewRequest("GET", "/info", http.NoBody)

		h.GetInfo(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response model.InfoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, uint64(1000), response.Coins)
		assert.Len(t, response.Inventory, 1)
		userService.AssertExpectations(t)
		transferService.AssertExpectations(t)
	})

	t.Run("пользователь не аутентифицирован", func(t *testing.T) {
		h := NewHandler(&mockUserService{}, &mockTransferService{}, &mockMerchService{})

		c, w := setupTestContext()
		c.Request = httptest.NewRequest("GET", "/info", http.NoBody)

		h.GetInfo(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestSendCoin(t *testing.T) {
	t.Run("успешная отправка монет", func(t *testing.T) {
		transferService := new(mockTransferService)
		h := NewHandler(&mockUserService{}, transferService, &mockMerchService{})

		transferService.On("SendCoins", mock.Anything, "sender", mock.AnythingOfType("string"), uint64(100)).Return(nil)

		c, w := setupTestContext()
		c.Set("username", "sender")
		body := bytes.NewBufferString(`{"to_user":"receiver","amount":100}`)
		c.Request = httptest.NewRequest("POST", "/sendCoin", body)
		c.Request.Header.Set("Content-Type", "application/json")

		h.SendCoin(c)

		assert.Equal(t, http.StatusOK, w.Code)
		transferService.AssertExpectations(t)
	})

	t.Run("недостаточно средств", func(t *testing.T) {
		transferService := new(mockTransferService)
		h := NewHandler(&mockUserService{}, transferService, &mockMerchService{})

		transferService.On("SendCoins", mock.Anything, "sender", mock.AnythingOfType("string"), uint64(1000)).
			Return(domain.ErrInsufficientFunds)

		c, w := setupTestContext()
		c.Set("username", "sender")
		body := bytes.NewBufferString(`{"to_user":"receiver","amount":1000}`)
		c.Request = httptest.NewRequest("POST", "/sendCoin", body)
		c.Request.Header.Set("Content-Type", "application/json")

		h.SendCoin(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		transferService.AssertExpectations(t)
	})
}

func TestBuyMerch(t *testing.T) {
	t.Run("успешная покупка", func(t *testing.T) {
		merchService := new(mockMerchService)
		h := NewHandler(&mockUserService{}, &mockTransferService{}, merchService)

		merchService.On("BuyMerch", mock.Anything, "buyer", "item1").Return(nil)

		c, w := setupTestContext()
		c.Set("username", "buyer")
		c.Params = []gin.Param{{Key: "item", Value: "item1"}}
		c.Request = httptest.NewRequest("GET", "/buy/item1", http.NoBody)

		h.BuyMerch(c)

		assert.Equal(t, http.StatusOK, w.Code)
		merchService.AssertExpectations(t)
	})

	t.Run("недостаточно средств для покупки", func(t *testing.T) {
		merchService := new(mockMerchService)
		h := NewHandler(&mockUserService{}, &mockTransferService{}, merchService)

		merchService.On("BuyMerch", mock.Anything, "buyer", "expensive-item").
			Return(domain.ErrInsufficientFunds)

		c, w := setupTestContext()
		c.Set("username", "buyer")
		c.Params = []gin.Param{{Key: "item", Value: "expensive-item"}}
		c.Request = httptest.NewRequest("GET", "/buy/expensive-item", http.NoBody)

		h.BuyMerch(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		merchService.AssertExpectations(t)
	})
}

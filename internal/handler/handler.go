package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/netscrawler/avito-shop/internal/domain"
	"github.com/netscrawler/avito-shop/internal/model"
	"github.com/netscrawler/avito-shop/internal/service"
)

// Коды ошибок
const (
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeInsufficientFunds  = "INSUFFICIENT_FUNDS"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeInternalError      = "INTERNAL_ERROR"
)

// Handler обрабатывает HTTP запросы
type Handler struct {
	userService     service.UserService
	transferService service.TransferService
	merchService    service.MerchService
}

// NewHandler создает новый экземпляр обработчика
func NewHandler(userService service.UserService, transferService service.TransferService, merchService service.MerchService) *Handler {
	return &Handler{
		userService:     userService,
		transferService: transferService,
		merchService:    merchService,
	}
}

// HealthCheck проверяет работоспособность сервиса
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Authenticate аутентифицирует пользователя
func (h *Handler) Authenticate(c *gin.Context) {
	var req model.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "Неверный формат запроса")
		return
	}

	token, err := h.userService.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			h.handleError(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, "Неверные учетные данные")
		default:
			h.handleError(c, http.StatusInternalServerError, ErrCodeInternalError, "Ошибка аутентификации")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetInfo возвращает информацию о пользователе
func (h *Handler) GetInfo(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		h.handleError(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, "Пользователь не аутентифицирован")
		return
	}

	user, err := h.userService.GetUserInfo(c.Request.Context(), username)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			h.handleError(c, http.StatusNotFound, ErrCodeNotFound, "Пользователь не найден")
		default:
			h.handleError(c, http.StatusInternalServerError, ErrCodeInternalError, "Ошибка получения информации")
		}
		return
	}

	transactions, err := h.transferService.GetTransactionHistory(c.Request.Context(), username)
	if err != nil {
		h.handleError(c, http.StatusInternalServerError, ErrCodeInternalError, "Ошибка получения истории транзакций")
		return
	}

	resp := model.InfoResponse{
		Coins:       user.Coins,
		Inventory:   user.Inventory,
		CoinHistory: transactions,
	}
	c.JSON(http.StatusOK, resp)
}

// SendCoin отправляет монеты другому пользователю
func (h *Handler) SendCoin(c *gin.Context) {
	var req model.SendCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "Неверный формат запроса")
		return
	}

	sender := c.GetString("username")
	if sender == "" {
		h.handleError(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, "Пользователь не аутентифицирован")
		return
	}

	err := h.transferService.SendCoins(c.Request.Context(), sender, req.ToUser, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientFunds):
			h.handleError(c, http.StatusBadRequest, ErrCodeInsufficientFunds, "Недостаточно средств")
		case errors.Is(err, domain.ErrUserNotFound):
			h.handleError(c, http.StatusNotFound, ErrCodeNotFound, "Получатель не найден")
		default:
			h.handleError(c, http.StatusInternalServerError, ErrCodeInternalError, "Ошибка перевода")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// BuyMerch обрабатывает покупку товара
func (h *Handler) BuyMerch(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		h.handleError(c, http.StatusUnauthorized, ErrCodeInvalidCredentials, "Пользователь не аутентифицирован")
		return
	}

	merchName := c.Param("item")
	if merchName == "" {
		h.handleError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "Не указан товар")
		return
	}

	err := h.merchService.BuyMerch(c.Request.Context(), username, merchName)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientFunds):
			h.handleError(c, http.StatusBadRequest, ErrCodeInsufficientFunds, "Недостаточно средств")
		case errors.Is(err, domain.ErrMerchNotFound):
			h.handleError(c, http.StatusNotFound, ErrCodeNotFound, "Товар не найден")
		default:
			h.handleError(c, http.StatusInternalServerError, ErrCodeInternalError, "Ошибка покупки")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleError обрабатывает ошибки и отправляет соответствующий ответ
func (h *Handler) handleError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"errors": code + ": " + message,
	})
}

type SendCoinRequest struct {
	ToUser string `json:"to_user" binding:"required"`
	Amount uint64 `json:"amount" binding:"required,gt=0"`
}

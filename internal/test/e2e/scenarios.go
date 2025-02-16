package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/netscrawler/avito-shop/internal/model"
)

// TestUserRegistrationAndAuthentication тестирует регистрацию и аутентификацию пользователя
func (s *E2ETestSuite) TestUserRegistrationAndAuthentication() error {
	// Подготовка данных для регистрации
	username := "test_user"
	password := "testpass123"

	// Регистрация пользователя
	token, err := s.registerAndLogin(username, password)
	if err != nil {
		return fmt.Errorf("failed to register and login user: %w", err)
	}
	if token == "" {
		return fmt.Errorf("received empty token")
	}

	// Проверка аутентификации с полученным токеном
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/info", s.baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	return nil
}

// TestCoinTransfer тестирует перевод монет между пользователями
func (s *E2ETestSuite) TestCoinTransfer() error {
	// Создаем двух пользователей
	sender := "test_sender"
	receiver := "test_receiver"
	password := "testpass123"

	// Регистрируем и авторизуем отправителя
	senderToken, err := s.registerAndLogin(sender, password)
	if err != nil {
		return fmt.Errorf("failed to register sender: %w", err)
	}

	// Регистрируем получателя
	_, err = s.registerAndLogin(receiver, password)
	if err != nil {
		return fmt.Errorf("failed to register receiver: %w", err)
	}

	// Отправляем монеты
	transferReq := model.SendCoinRequest{
		ToUser: receiver,
		Amount: 100,
	}

	reqBody, err := json.Marshal(transferReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/api/sendCoin", s.baseURL),
		bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", senderToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	// Проверяем баланс отправителя
	req, err = http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/info", s.baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", senderToken))

	resp, err = s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	var userInfo model.InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if userInfo.Coins != 900 { // Начальный баланс 1000 - 100
		return fmt.Errorf("expected sender balance 900, got %d", userInfo.Coins)
	}

	return nil
}

// TestMerchPurchase тестирует покупку товара
func (s *E2ETestSuite) TestMerchPurchase() error {
	// Регистрируем пользователя
	username := "test_buyer"
	password := "testpass123"

	token, err := s.registerAndLogin(username, password)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	// Покупаем мерч
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/buy/cup", s.baseURL),
		nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	// Проверяем инвентарь пользователя
	req, err = http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/info", s.baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err = s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	var userInfo model.InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Проверяем, что товар появился в инвентаре
	found := false
	for _, item := range userInfo.Inventory {
		if item.Type == "cup" && item.Quantity > 0 {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cup not found in inventory")
	}

	// Проверяем, что с баланса списались монеты
	if userInfo.Coins != 980 { // 1000 - 20 (цена чашки)
		return fmt.Errorf("expected coins to be 980, got %d", userInfo.Coins)
	}

	return nil
}

// TestTransactionHistory тестирует историю транзакций
func (s *E2ETestSuite) TestTransactionHistory() error {
	// Регистрируем пользователя
	username := "test_history_user"
	password := "testpass123"

	token, err := s.registerAndLogin(username, password)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	// Регистрируем второго пользователя для перевода
	receiverUsername := "test_receiver"
	receiverPassword := "testpass123"

	_, err = s.registerAndLogin(receiverUsername, receiverPassword)
	if err != nil {
		return fmt.Errorf("failed to register receiver: %w", err)
	}

	// Создаем транзакцию перевода монет
	sendReq, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/api/sendCoin", s.baseURL),
		strings.NewReader(`{"toUser": "test_receiver", "amount": 100}`))
	if err != nil {
		return fmt.Errorf("failed to create send request: %w", err)
	}
	sendReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	sendReq.Header.Set("Content-Type", "application/json")

	sendResp, err := s.httpClient.Do(sendReq)
	if err != nil {
		return fmt.Errorf("failed to send coins: %w", err)
	}
	defer sendResp.Body.Close()

	if sendResp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK for send coins, got %d", sendResp.StatusCode)
	}

	// Получаем историю транзакций
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/info", s.baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	var userInfo model.InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(userInfo.CoinHistory.Sent) == 0 && len(userInfo.CoinHistory.Received) == 0 {
		return fmt.Errorf("expected non-empty transaction history")
	}

	return nil
}

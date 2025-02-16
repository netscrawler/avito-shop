package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/netscrawler/avito-shop/internal/model"
)

// makeRequest выполняет HTTP запрос с заданными параметрами
func (s *E2ETestSuite) makeRequest(method, path string, body interface{}, token string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", s.baseURL, path), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return s.httpClient.Do(req)
}

// registerAndLogin регистрирует нового пользователя и возвращает токен аутентификации
func (s *E2ETestSuite) registerAndLogin(username, password string) (string, error) {
	// Регистрация и аутентификация пользователя через единый эндпоинт
	loginReq := model.AuthRequest{
		Username: username,
		Password: password,
	}

	resp, err := s.makeRequest(http.MethodPost, "/api/auth", loginReq, "")
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate user: status code %d", resp.StatusCode)
	}

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.Token, nil
}

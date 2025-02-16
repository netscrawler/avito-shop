package e2e

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netscrawler/avito-shop/internal/config"
)

// E2ETestSuite представляет набор e2e тестов
type E2ETestSuite struct {
	cfg        *config.Config
	dbPool     *pgxpool.Pool
	httpClient *http.Client
	baseURL    string
}

// NewE2ETestSuite создает новый экземпляр тестового набора
func NewE2ETestSuite(cfg *config.Config, dbPool *pgxpool.Pool, httpClient *http.Client, baseURL string) *E2ETestSuite {
	return &E2ETestSuite{
		cfg:        cfg,
		dbPool:     dbPool,
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// Run запускает все тесты
func (s *E2ETestSuite) Run() error {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"TestUserRegistrationAndAuthentication", s.TestUserRegistrationAndAuthentication},
		{"TestCoinTransfer", s.TestCoinTransfer},
		{"TestMerchPurchase", s.TestMerchPurchase},
		{"TestTransactionHistory", s.TestTransactionHistory},
	}

	for _, test := range tests {
		fmt.Printf("Running test: %s\n", test.name)
		if err := test.fn(); err != nil {
			return fmt.Errorf("%s failed: %w", test.name, err)
		}
		fmt.Printf("Test passed: %s\n", test.name)
	}

	return nil
}

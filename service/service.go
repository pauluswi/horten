package service

import (
	"fmt"
	"sync"
	"time"
)

// Account represents account details.
type Account struct {
	AccountNumber string  `json:"accountNumber"`
	CustomerName  string  `json:"customerName"`
	Balance       float64 `json:"balance"`
}

// Mock data for accounts
var accounts = map[string]Account{
	"123456": {AccountNumber: "123456", CustomerName: "John Doe", Balance: 1000.0},
	"654321": {AccountNumber: "654321", CustomerName: "Jane Doe", Balance: 2000.0},
}

// AccountService handles business logic related to accounts.
type AccountService struct {
	mutex sync.Mutex
}

// NewAccountService creates a new AccountService.
func NewAccountService() *AccountService {
	return &AccountService{}
}

// GetBalance fetches the balance for a given account number.
func (s *AccountService) GetBalance(accountNumber string) (Account, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if account, exists := accounts[accountNumber]; exists {
		// Simulate concurrency with artificial delay
		time.Sleep(100 * time.Millisecond)
		return account, nil
	}
	return Account{}, fmt.Errorf("account not found")
}

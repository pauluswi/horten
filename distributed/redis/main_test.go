package main

import (
	"sync"
	"testing"
	"time"
)

// MockRedisClient simulates a Redis client for testing
type MockRedisClient struct {
	locks map[string]string
	mu    sync.Mutex
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{locks: make(map[string]string)}
}

func (m *MockRedisClient) SetNX(key, value string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.locks[key]; exists {
		return false, nil
	}

	m.locks[key] = value
	go func() {
		time.Sleep(ttl)
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.locks[key] == value {
			delete(m.locks, key)
		}
	}()

	return true, nil
}

func (m *MockRedisClient) Get(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, exists := m.locks[key]
	if !exists {
		return "", nil
	}
	return value, nil
}

func (m *MockRedisClient) Del(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.locks[key]; exists {
		delete(m.locks, key)
		return 1, nil
	}
	return 0, nil
}

func TestProcessTransactionWithMockRedis(t *testing.T) {
	// Initialize mock Redis client
	rdb := NewMockRedisClient()

	// Setup accounts
	accounts := map[string]*Account{
		"11111": {AccountNumber: "11111", Balance: 1000},
		"22222": {AccountNumber: "22222", Balance: 2000},
	}

	// List of transactions
	transactions := []Transaction{
		{"11111", -200},  // Debit
		{"11111", 300},   // Credit
		{"22222", -500},  // Debit
		{"22222", -3000}, // Insufficient funds
		{"11111", 100},   // Credit
	}

	// Expected final balances
	expectedBalances := map[string]float64{
		"11111": 1200,
		"22222": 1500,
	}

	var wg sync.WaitGroup

	// Process transactions concurrently
	for _, tr := range transactions { // Rename loop variable to 'tr'
		account, exists := accounts[tr.AccountNumber]
		if !exists {
			t.Errorf("Account %s not found", tr.AccountNumber) // Use correct testing object
			continue
		}

		wg.Add(1)
		go func(acc *Account, tr Transaction) {
			defer wg.Done()
			ProcessTransaction(acc, tr, rdb, &wg)
		}(account, tr)
	}

	wg.Wait() // Wait for all transactions to complete

	// Verify final balances
	for accountNumber, account := range accounts {
		if account.Balance != expectedBalances[accountNumber] {
			t.Errorf("Account %s: expected balance %.2f, got %.2f",
				account.AccountNumber, expectedBalances[accountNumber], account.Balance)
		}
	}
}

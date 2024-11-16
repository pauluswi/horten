package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var ctx = context.Background()

// MockRedis simulates a Redis client with basic lock functionality
type MockRedis struct {
	data map[string]string
	mu   sync.Mutex
}

// NewMockRedis creates a new MockRedis instance
func NewMockRedis() *MockRedis {
	return &MockRedis{
		data: make(map[string]string),
	}
}

// SetNX simulates the Redis SETNX command
func (r *MockRedis) SetNX(key, value string, ttl time.Duration) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[key]; exists {
		return false, nil
	}

	// Simulate setting the value with a TTL
	r.data[key] = value
	go func() {
		time.Sleep(ttl)
		r.mu.Lock()
		defer r.mu.Unlock()
		// Only delete if the same value still exists (avoid deleting newer locks)
		if r.data[key] == value {
			delete(r.data, key)
		}
	}()

	return true, nil
}

// Get simulates the Redis GET command
func (r *MockRedis) Get(key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value, exists := r.data[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}
	return value, nil
}

// Del simulates the Redis DEL command
func (r *MockRedis) Del(key string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.data[key]
	if !exists {
		return 0, nil
	}

	delete(r.data, key)
	return 1, nil
}

// Account represents a bank account
type Account struct {
	AccountNumber string
	Balance       float64
}

// Transaction represents a financial transaction
type Transaction struct {
	AccountNumber string
	Amount        float64 // Positive for credit, negative for debit
}

// RedisLock represents a distributed lock

type RedisLock struct {
	client *MockRedis
	key    string
	value  string
}

// AcquireLock tries to acquire the lock
func (lock *RedisLock) AcquireLock(ttl time.Duration) (bool, error) {
	return lock.client.SetNX(lock.key, lock.value, ttl)
}

// ReleaseLock releases the lock
func (lock *RedisLock) ReleaseLock() error {
	val, err := lock.client.Get(lock.key)
	if err != nil {
		return fmt.Errorf("lock not found")
	}

	// Ensure the lock is released by the process that acquired it
	if val == lock.value {
		_, err = lock.client.Del(lock.key)
		if err != nil {
			return err
		}
	}
	return nil
}

// ProcessTransaction processes a single transaction on an account with distributed locking
func ProcessTransaction(account *Account, transaction Transaction, rdb *MockRedis, wg *sync.WaitGroup) {
	defer wg.Done()

	lock := RedisLock{
		client: rdb,
		key:    fmt.Sprintf("account:%s:lock", account.AccountNumber),
		value:  fmt.Sprintf("unique-identifier-%d", time.Now().UnixNano()), // Use a unique identifier
	}

	// Retry logic for acquiring lock
	for i := 0; i < 3; i++ {
		acquired, err := lock.AcquireLock(5 * time.Second)
		if err != nil {
			fmt.Printf("Error acquiring lock for account %s: %v\n", account.AccountNumber, err)
			return
		}

		if acquired {
			break
		}

		if i == 2 {
			fmt.Printf("Could not acquire lock for account %s after retries. Transaction skipped.\n", account.AccountNumber)
			return
		}

		time.Sleep(100 * time.Millisecond) // Small delay before retry
	}

	// Lock acquired, process the transaction
	defer lock.ReleaseLock()

	if transaction.Amount < 0 && account.Balance+transaction.Amount < 0 {
		fmt.Printf("Insufficient funds for account %s. Transaction skipped.\n", account.AccountNumber)
		return
	}

	account.Balance += transaction.Amount
	fmt.Printf("Processed transaction of %.2f on account %s. New balance: %.2f\n",
		transaction.Amount, account.AccountNumber, account.Balance)
}

func main() {
	// Create a mock Redis client
	rdb := NewMockRedis()

	// Create accounts
	accounts := map[string]*Account{
		"12345": {AccountNumber: "12345", Balance: 1000},
		"67890": {AccountNumber: "67890", Balance: 2000},
	}

	// List of transactions
	transactions := []Transaction{
		{"12345", -200},  // Debit
		{"12345", 300},   // Credit
		{"67890", -500},  // Debit
		{"67890", -3000}, // Insufficient funds
		{"12345", 100},   // Credit
	}

	var wg sync.WaitGroup

	// Process transactions concurrently
	for _, t := range transactions {
		account, exists := accounts[t.AccountNumber]
		if !exists {
			fmt.Printf("Account %s not found\n", t.AccountNumber)
			continue
		}

		wg.Add(1)
		go ProcessTransaction(account, t, rdb, &wg)
	}

	wg.Wait() // Wait for all transactions to complete

	// Print final balances
	fmt.Println("\nFinal Account Balances:")
	for _, account := range accounts {
		fmt.Printf("Account %s: %.2f\n", account.AccountNumber, account.Balance)
	}
}

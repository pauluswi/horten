package main

import (
	"sync"
	"testing"
	"time"
)

func TestProcessTransactionWithMockRedis(t *testing.T) {
	// Initialize mock Redis client
	rdb := NewMockRedis()

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
	for _, tr := range transactions {
		tr := tr // Capture range variable
		account, exists := accounts[tr.AccountNumber]
		if !exists {
			t.Errorf("Account %s not found", tr.AccountNumber)
			continue
		}

		wg.Add(1)
		go ProcessTransaction(account, tr, rdb, &wg)
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

func TestRedisLock(t *testing.T) {
	// Initialize mock Redis client
	rdb := NewMockRedis()
	lock := RedisLock{
		client: rdb,
		key:    "test-key",
		value:  "test-value",
	}

	// Acquire lock
	success, err := lock.AcquireLock(1 * time.Second)
	if err != nil || !success {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Attempt to acquire the same lock again (should fail)
	success, err = lock.AcquireLock(1 * time.Second)
	if err != nil {
		t.Fatalf("Unexpected error acquiring lock: %v", err)
	}
	if success {
		t.Fatalf("Expected lock acquisition to fail, but it succeeded")
	}

	// Release lock
	if err := lock.ReleaseLock(); err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Acquire lock again (should succeed)
	success, err = lock.AcquireLock(1 * time.Second)
	if err != nil || !success {
		t.Fatalf("Failed to acquire lock after release: %v", err)
	}
}

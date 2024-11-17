package main

import (
	"sync"
	"testing"
)

func TestProcessTransaction(t *testing.T) {
	// Setup accounts
	accounts := map[string]*Account{
		"11111": {AccountNumber: "11111", Balance: 10000},
		"22222": {AccountNumber: "22222", Balance: 20000},
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
		"11111": 10200,
		"22222": 16500,
	}

	var wg sync.WaitGroup

	// Process transactions concurrently
	for _, tr := range transactions {
		tr := tr // Capture range variable for goroutine
		account, exists := accounts[tr.AccountNumber]
		if !exists {
			t.Errorf("Account %s not found", tr.AccountNumber)
			continue
		}

		wg.Add(1)
		go func(acc *Account, amount float64) {
			defer wg.Done() // Ensure Done is called even if ProcessTransaction fails
			acc.mu.Lock()
			if amount < 0 && acc.Balance+amount < 0 {
				acc.mu.Unlock()
				t.Errorf("Insufficient funds for account %s", acc.AccountNumber)
				return
			}
			acc.Balance += amount
			acc.mu.Unlock()
		}(account, tr.Amount)
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

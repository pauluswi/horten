package main

import (
	"fmt"
	"sync"
)

// Account represents a bank account
type Account struct {
	AccountNumber string
	Balance       float64
	mu            sync.Mutex // Mutex to protect the balance
}

// Transaction represents a financial transaction
type Transaction struct {
	AccountNumber string
	Amount        float64 // Positive for credit, negative for debit
}

// ProcessTransaction processes a single transaction on the account
func (a *Account) ProcessTransaction(amount float64, wg *sync.WaitGroup) {
	defer wg.Done() // Notify when the goroutine is finished

	// Lock the account to prevent race conditions
	a.mu.Lock()
	defer a.mu.Unlock()

	if amount < 0 && a.Balance+amount < 0 {
		fmt.Printf("Insufficient funds for account %s\n", a.AccountNumber)
		return
	}

	a.Balance += amount
	fmt.Printf("Processed transaction of %.2f on account %s. New balance: %.2f\n",
		amount, a.AccountNumber, a.Balance)
}

func main() {
	// Create accounts
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

	var wg sync.WaitGroup

	// Process transactions concurrently
	for _, t := range transactions {
		account, exists := accounts[t.AccountNumber]
		if !exists {
			fmt.Printf("Account %s not found\n", t.AccountNumber)
			continue
		}

		wg.Add(1)
		go account.ProcessTransaction(t.Amount, &wg)
	}

	wg.Wait() // Wait for all transactions to complete

	// Print final balances
	fmt.Println("\nFinal Account Balances:")
	for _, account := range accounts {
		fmt.Printf("Account %s: %.2f\n", account.AccountNumber, account.Balance)
	}
}

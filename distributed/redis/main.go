package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

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
	client RedisClient
	key    string
	value  string
}

// RedisClient defines the interface for a Redis client
type RedisClient interface {
	SetNX(key, value string, ttl time.Duration) (bool, error)
	Get(key string) (string, error)
	Del(key string) (int64, error)
}

// RedisAdapter wraps *redis.Client to implement RedisClient
type RedisAdapter struct {
	client *redis.Client
}

func (r *RedisAdapter) SetNX(key, value string, ttl time.Duration) (bool, error) {
	result, err := r.client.SetNX(ctx, key, value, ttl).Result()
	return result, err
}

func (r *RedisAdapter) Get(key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	return result, err
}

func (r *RedisAdapter) Del(key string) (int64, error) {
	result, err := r.client.Del(ctx, key).Result()
	return result, err
}

// AcquireLock tries to acquire the lock
func (lock *RedisLock) AcquireLock(ttl time.Duration) (bool, error) {
	success, err := lock.client.SetNX(lock.key, lock.value, ttl)
	if err != nil {
		return false, err
	}
	return success, nil
}

// ReleaseLock releases the lock
func (lock *RedisLock) ReleaseLock() error {
	val, err := lock.client.Get(lock.key)
	if err == redis.Nil {
		return fmt.Errorf("lock not found")
	} else if err != nil {
		return err
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
func ProcessTransaction(account *Account, transaction Transaction, rdb RedisClient, wg *sync.WaitGroup) {
	defer wg.Done()

	lock := RedisLock{
		client: rdb,
		key:    fmt.Sprintf("account:%s:lock", account.AccountNumber),
		value:  "unique-identifier", // Use a UUID in production
	}

	// Try to acquire the lock
	acquired, err := lock.AcquireLock(5 * time.Second)
	if err != nil {
		fmt.Printf("Error acquiring lock for account %s: %v\n", account.AccountNumber, err)
		return
	}

	if !acquired {
		fmt.Printf("Could not acquire lock for account %s. Transaction skipped.\n", account.AccountNumber)
		return
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
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Adjust based on your setup
		DB:   0,
	})

	// Wrap Redis client with adapter
	rdb := &RedisAdapter{client: redisClient}

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
		go ProcessTransaction(account, t, rdb, &wg)
	}

	wg.Wait() // Wait for all transactions to complete

	// Print final balances
	fmt.Println("\nFinal Account Balances:")
	for _, account := range accounts {
		fmt.Printf("Account %s: %.2f\n", account.AccountNumber, account.Balance)
	}
}

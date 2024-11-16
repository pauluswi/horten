
# Concurrency Handling in Golang

![Go Version](https://img.shields.io/github/go-mod/go-version/pauluswi/horten)
![License](https://img.shields.io/github/license/pauluswi/horten)
![Build Status](https://img.shields.io/github/workflow/status/pauluswi/horten/CI)


This repository shows two approaches to handling concurrency in Golang:

1. **Single Process Lock using `sync.Mutex`**
2. **Distributed Lock using Redis Locks**

The examples simulate a bank transaction system where multiple transactions are processed concurrently while ensuring data consistency.

---

## 1: Single Process Lock with `sync.Mutex`

### **Overview**
This approach uses Golang's `sync.Mutex` to enforce mutual exclusion within a single process. It ensures that only one goroutine can access and modify shared resources (e.g., account balance) at a time.

### **Code Highlights**
- **Use of Mutex**: A `sync.Mutex` instance is used to lock critical sections.
- **Concurrency**: Goroutines process transactions concurrently but safely.
- **Example Use Case**:
  - Suitable for single-machine systems where all operations are within the same memory space.

### **Advantages**
- Simple and lightweight.
- No external dependencies.
- Very fast as operations are performed in memory.

### **Limitations**
- Cannot be used in distributed systems.
- Limited to processes running on a single machine.

### **Code**
- Use WaitGroup
```go
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
```

- Use 
```go
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

```

### Output
```
     Single Machine Locks Output:

    Processed transaction of -200.00 on account 11111. New balance: 800.00
    Processed transaction of 100.00 on account 11111. New balance: 900.00
    Processed transaction of 300.00 on account 11111. New balance: 1200.00
    Processed transaction of -500.00 on account 22222. New balance: 1500.00
    Insufficient funds for account 22222
    
    Final Account Balances:
    Account 11111: 1200.00
    Account 22222: 1500.00
```

---

## 2: Distributed Lock using Redis Locks

### **Overview**
This approach uses Redis to simulate distributed locking. Each critical section is protected by a distributed lock, ensuring data consistency across multiple processes or machines.

### **Code Highlights**
- **Mock Redis**: A mock Redis client is used for local testing.
- **Distributed Lock**: Implemented using Redis commands (`SETNX`, `GET`, `DEL`) to acquire and release locks.
- **Example Use Case**:
  - Ideal for distributed systems where multiple instances or processes need to synchronize access to shared resources.

### **Advantages**
- Supports distributed systems.
- Can synchronize processes across multiple machines.
- Scales well with cloud-native environments (e.g., Kubernetes).

### **Limitations**
- Requires a running Redis instance or equivalent service.
- Slightly slower due to network latency.
- More complex to implement and maintain.

### **Mock**
- To make a simple ready-to-go process, I utilize mock for redis to simulate redis without run real Redis instance.

### Output
```
    Distributed Locks Output:
    
    Processed transaction of 100.00 on account 11111. New balance: 1100.00
    Insufficient funds for account 22222. Transaction skipped.
    Processed transaction of 300.00 on account 11111. New balance: 1400.00
    Processed transaction of -500.00 on account 22222. New balance: 1500.00
    Processed transaction of -200.00 on account 11111. New balance: 1200.00
    
    Final Account Balances:
    Account 11111: 1200.00
    Account 22222: 1500.00
```

---

## Comparison: `sync.Mutex` vs Redis Locks

| Feature                | `sync.Mutex`                         | Redis Locks                             |
|------------------------|---------------------------------------|-----------------------------------------|
| **Scope**             | Single process                       | Distributed processes                   |
| **Performance**       | Very fast (in-memory)                | Slightly slower (network I/O overhead)  |
| **Complexity**        | Simple                               | Medium                                  |
| **Dependencies**      | None                                 | Redis or equivalent                     |
| **Use Case**          | Single-machine systems               | Multi-machine or distributed systems    |
| **Failure Handling**  | Handled within process               | Requires TTL and retries                |

---

## When to Use

### Use `sync.Mutex` when:
- Your application runs on a single machine.
- You don’t need to synchronize across multiple processes or machines.
- Simplicity and performance are priorities.

### Use Redis Locks when:
- Your application runs in a distributed system.
- You need to synchronize access across multiple processes or machines.
- Scalability and cross-instance consistency are required.

---

## How to Run

1. **Single Process Lock (Mutex)**
   - Clone the repository.
   - Navigate to the `mutex` example folder.
   - Run the program:
     ```bash
     go run main.go
     ```

2. **Distributed Lock (Redis)**
   - Ensure Redis is running or use the mock Redis provided.
   - Navigate to the `redis-lock` example folder.
   - Run the program:
     ```bash
     go run main.go
     ```

---

## License
This project is licensed under the MIT License.









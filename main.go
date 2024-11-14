package main

import (
	"context"
	"fmt"

	"horten/config"
	"horten/handler"
	"horten/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	//"github.com/go-delve/delve/pkg/config"
)

func main() {
	// Load configuration
	config := config.LoadConfig()

	// Initialize services
	accountService := service.NewAccountService()
	handler := handler.NewHandler(accountService)

	// Create HTTP server
	http.Handle("/account/balance", handler.Logger(http.HandlerFunc(handler.GetAccountBalanceHandler)))

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", config.Port),
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server is running on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", config.Port, err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

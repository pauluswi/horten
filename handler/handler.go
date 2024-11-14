package handler

import (
	"encoding/json"
	"horten/service"
	"log"
	"net/http"
)

// Handler wraps dependencies for HTTP handlers.
type Handler struct {
	accountService *service.AccountService
}

// NewHandler creates a new Handler.
func NewHandler(accountService *service.AccountService) *Handler {
	return &Handler{accountService: accountService}
}

// GetAccountBalanceHandler handles requests for account balances.
func (h *Handler) GetAccountBalanceHandler(w http.ResponseWriter, r *http.Request) {
	accountNumber := r.URL.Query().Get("accountNumber")
	if accountNumber == "" {
		http.Error(w, "accountNumber is required", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.GetBalance(accountNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response, _ := json.Marshal(account)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// Logger middleware for logging HTTP requests.
func (h *Handler) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

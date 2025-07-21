package handler

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
)

var (
	urlStore = make(map[string]string)
	mu       sync.Mutex
	letters  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func generateCode(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Generate a short code and store the mapping
	code := generateCode(6)
	mu.Lock()
	urlStore[code] = req.URL
	mu.Unlock()

	// Return the short URL (update with your actual domain)
	shortURL := "https://" + r.Host + "/api/shorten/" + code
	resp := shortenResponse{ShortURL: shortURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

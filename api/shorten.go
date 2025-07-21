package handler

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rdb     *redis.Client
)

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // e.g. "us1-xxx.upstash.io:6379"
		Password: os.Getenv("REDIS_PASSWORD"), // e.g. "yourpassword"
	})
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

// Handler handles POST /api/shorten for Vercel serverless
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

	code := generateCode(6)
	ctx := context.Background()
	if err := rdb.Set(ctx, code, req.URL, 0).Err(); err != nil {
		log.Println("Redis SET error:", err)
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	shortURL := "https://" + r.Host + "/" + code
	resp := shortenResponse{ShortURL: shortURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func generateCode(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

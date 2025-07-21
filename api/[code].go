package handler

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

var rdbRoot *redis.Client

func init() {
	rdbRoot = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
}

// Handler handles GET /{code} for Vercel serverless
func Handler(w http.ResponseWriter, r *http.Request) {
	// Extract code from the URL path (should be /{code})
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	code := parts[1]

	ctx := context.Background()
	longURL, err := rdbRoot.Get(ctx, code).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

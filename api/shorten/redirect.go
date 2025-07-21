package handler

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
}

// Handler handles GET /api/shorten/{code} for Vercel serverless
// Vercel expects an exported Handler symbol
func Handler(w http.ResponseWriter, r *http.Request) {
	// Extract code from the URL path (should be /api/shorten/{code})
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	code := parts[len(parts)-1]

	ctx := context.Background()
	longURL, err := rdb.Get(ctx, code).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

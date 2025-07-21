package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

var (
	urlStore = make(map[string]string)
	storeMu  sync.RWMutex
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

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Generate unique code
	var code string
	for {
		code = generateCode(6)
		storeMu.RLock()
		_, exists := urlStore[code]
		storeMu.RUnlock()
		if !exists {
			break
		}
	}

	storeMu.Lock()
	urlStore[code] = req.URL
	storeMu.Unlock()

	resp := shortenResponse{ShortURL: "http://localhost:8080/" + code}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" {
		http.NotFound(w, r)
		return
	}
	storeMu.RLock()
	longURL, exists := urlStore[code]
	storeMu.RUnlock()
	if !exists {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {
	http.HandleFunc("/shorten", shortenHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		// If the path starts with /static/, serve static files
		if len(r.URL.Path) > 8 && r.URL.Path[:8] == "/static/" {
			http.ServeFile(w, r, r.URL.Path[1:])
			return
		}
		// Otherwise, treat as a short code
		redirectHandler(w, r)
	})

	log.Println("URL shortener running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

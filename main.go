package main

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode(n int) string {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err) // prod'da logla + 500 dön
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

type shortenRequest struct {
	URL string `json:"url"`
}
type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

var (
	urlStore = make(map[string]string)
	storeMu  sync.RWMutex
)

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// unique code
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

	resp := shortenResponse{ShortURL: "https://" + r.Host + "/" + code}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	storeMu.RLock()
	longURL, ok := urlStore[code]
	storeMu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, longURL, http.StatusFound)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "static/index.html")
		return
	}
	redirectHandler(w, r)
}

func main() {
	mux := http.NewServeMux()

	// API
	mux.Handle("/shorten", cors(http.HandlerFunc(shortenHandler)))

	// Statik dosyalar
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", cors(http.StripPrefix("/static/", fs)))

	// Kök + redirect
	mux.Handle("/", cors(http.HandlerFunc(rootHandler)))

	log.Println("HTTPS up on https://localhost:8080")
	log.Fatal(http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", mux))
}

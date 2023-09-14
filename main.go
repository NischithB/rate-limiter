package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type RequestCounter struct {
	id    string
	count int
}

var cache = map[string]*RequestCounter{}

func RateLimiter(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adr := r.RemoteAddr
		ip := strings.Split(adr, ":")[0]
		ctr, exists := cache[ip]
		if !exists {
			cache[ip] = &RequestCounter{id: ip, count: 0}
			ctr = cache[ip]
			time.AfterFunc(time.Second*time.Duration(10), func() { delete(cache, ip) })
		}
		ctr.count++
		if ctr.count > 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		fn(w, r)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Get("/", RateLimiter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Rate Limiter is getting ready!!"))
	}))

	addr := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	server := &http.Server{Addr: addr, Handler: router}
	log.Printf("Server is available at: http://%s", addr)
	log.Fatal(server.ListenAndServe())
}

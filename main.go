package main
import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)
type Client struct {
	requests    int
	lastRequest time.Time
}
var clients = make(map[string]*Client)
var mutex sync.Mutex
const maxRequests = 5
const timeWindow = time.Minute

// ================= RATE LIMITER =================

func rateLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get only IP address (remove port)
		ip := strings.Split(r.RemoteAddr, ":")[0]
		mutex.Lock()
		defer mutex.Unlock()
		client, exists := clients[ip]
		// First request from this client
		if !exists {
			clients[ip] = &Client{
				requests:    1,
				lastRequest: time.Now(),
			}
			next(w, r)
			return
		}
		// Reset after time window
		if time.Since(client.lastRequest) > timeWindow {
			client.requests = 1
			client.lastRequest = time.Now()
			next(w, r)
			return
		}
		// Limit exceeded
		if client.requests >= maxRequests {
			http.Error(
				w,
				"Rate limit exceeded",
				http.StatusTooManyRequests,
			)
			return
		}
		// Increase request count
		client.requests++
		next(w, r)

	}

}
// ================= HANDLER =================
func homeHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	fmt.Fprintln(
		w,
		"Request successful",
	)

}
// ================= MAIN =================

func main() {
	http.HandleFunc(
		"/",
		rateLimiter(homeHandler),
	)
	fmt.Println(
		"Server running on port 8080",
	)
	log.Fatal(
		http.ListenAndServe(
			":8080",
			nil,
		),
	)

}
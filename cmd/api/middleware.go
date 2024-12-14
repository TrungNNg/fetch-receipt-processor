package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// recoverPanic middleware send a JSON response to client instead of the default response.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If next handler panic, this defer function will run when Go unwind the call stack.
		defer func() {
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response which will trigger
				// Go's HTTP server to automatically close the current connection
				// after a response has been sent.
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// rateLimit is a middleware function that implements rate limiting for HTTP requests based on the client's IP address.
// It uses a background goroutine to periodically clean up expired client entries from the rate limit map.
//
// The function works as follows:
//   - For each incoming request, it checks whether the client's IP address has been rate-limited based on the configured
//     requests per second (rps) and burst limit.
//   - If the rate limit has been exceeded, it responds with a rate limit exceeded message.
//   - Otherwise, it allows the request to proceed to the next handler in the chain.
//
// Additionally, the function manages rate-limiting state in memory for each client and cleans up entries that have been
// inactive for more than three minutes to conserve memory. The rate-limiting state is stored in a map, with each entry
// containing a limiter and the last time the client was seen.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Define a client struct to hold the rate limiter and last seen time for each
	// client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex
		// Update the map so the values are pointers to a client struct.
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map once
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	// Closure capture mu sync.Mutex and clients map[string]*client which are
	// reference types which refer to the same mutex and clients map for
	// all requests.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only carry out the check if rate limiting is enabled.
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

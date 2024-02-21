package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// preiodically remove old entries from the map to prevent memory bottleneck
	// background goroutine
	go func() {
		time.Sleep(time.Minute) // runs every min
		mu.Lock()
		for ip, client := range clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}()

	// limiter := rate.NewLimiter(2, 4) // avg 2 reqs/s and burst 4 reqs. Refills at 1 req per 2 seconds (1/r)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
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
	
			mu.Unlock() // important to not defer this as if not it will only run after serving http (downstream handlers)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// deferred funcs are always called when go unwinds the stack from a panic
		defer func() {
			if err := recover(); err != nil { // recover is a built-in func to check if panic has occurred
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("error %s", err)) // since err is of type any
			}
		}()

		next.ServeHTTP(w, r) // I believe continue receiving API reqs, not just stopping
	})
}

// Only panics for this goroutine is handled by this middleware, if we spin up other goroutines,
// we have to make sure the panics are handled for those goroutines too

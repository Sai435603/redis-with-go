Go
package main

import (
	"net/http"
)

type Limiter interface {
	Allow() bool
}

func RateLimitMiddleware(limiter Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		// Proceed to the actual handler
		next.ServeHTTP(w, r)
	})
}
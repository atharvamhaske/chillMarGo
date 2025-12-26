package main

import (
	"log"
	"net/http"
	"github.com/chillMarGO/internal/handlers"
	"github.com/chillMarGO/internal/middleware"
	ratelimiter "github.com/chillMarGO/internal/rate-limiter"
)

func main() {
	l := ratelimiter.NewLimiter(10, 1) 

	tokenBucketAlgo := ratelimiter.TokenBucketAlgo(l)
	rateLimited := middleware.RateLimiterWrapper(handlers.Resource, tokenBucketAlgo)

	// route
	http.HandleFunc("/v1/resource", rateLimited)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
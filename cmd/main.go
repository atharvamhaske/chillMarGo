package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chillMarGO/internal/handlers"
	"github.com/chillMarGO/internal/middleware"
	ratelimiter "github.com/chillMarGO/internal/rate-limiter"
)

var port = ":8080"

func main() {
	l := ratelimiter.NewLimiter(10, 1)

	tokenBucketAlgo := ratelimiter.TokenBucketAlgo(l)
	rateLimited := middleware.RateLimiterWrapper(handlers.Resource, tokenBucketAlgo)

	// Rate limited route
	http.HandleFunc("/v1/resource", rateLimited)

	fmt.Printf("Server running on http://localhost%s\n", port)

	log.Fatal(http.ListenAndServe(port, nil))
}

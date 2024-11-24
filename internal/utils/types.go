package utils

import (
	"net/http"
)

// ClientWrapper represents a wrapped HTTP client with rate limiting capabilities
type ClientWrapper struct {
	Client      *http.Client
	RateLimiter *RateLimiter
	IsDirect    bool
}

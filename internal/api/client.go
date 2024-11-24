package api

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"api-proxy/internal/config"
	"api-proxy/internal/utils"
	"github.com/sirupsen/logrus"
)

type Client struct {
	config       *config.Config
	logger       *logrus.Logger
	rateLimiters *utils.RateLimiterManager
	keyIndex     atomic.Int32
	proxyIndex   atomic.Int32
	clients      []utils.ClientWrapper
	mu           sync.RWMutex
	apiKeys      map[string]int
}

func NewClient(cfg *config.Config, logger *logrus.Logger) *Client {
	apiKeys := make(map[string]int)
	for _, keyLimit := range cfg.Values {
		for key, limit := range keyLimit {
			apiKeys[key] = limit
		}
	}

	c := &Client{
		config:       cfg,
		logger:       logger,
		rateLimiters: utils.NewRateLimiterManager(),
		clients:      make([]utils.ClientWrapper, 0),
		apiKeys:      apiKeys,
	}

	if cfg.DirectAccess {
		c.clients = append(c.clients, utils.ClientWrapper{
			Client: &http.Client{
				Timeout: time.Second * 30,
			},
			RateLimiter: utils.NewRateLimiter(cfg.DirectAccessTPS),
			IsDirect:    true,
		})
		logger.Infof("Added direct access client with TPS limit: %d", cfg.DirectAccessTPS)
	}

	for _, proxyURL := range cfg.Proxies {
		client, err := utils.CreateProxyClient(proxyURL)
		if err != nil {
			logger.Errorf("Failed to create proxy client for %s: %v", proxyURL, err)
			continue
		}
		c.clients = append(c.clients, utils.ClientWrapper{
			Client:      client,
			RateLimiter: utils.NewRateLimiter(cfg.ProxyTPS), // Add rate limiting for proxy clients
			IsDirect:    false,
		})
		logger.Infof("Added proxy client: %s", proxyURL)
	}

	if len(c.clients) == 0 {
		logger.Fatal("No working clients available (neither direct access nor proxies)")
	}

	return c
}

func (c *Client) getNextAPIKey() (string, int, error) {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if key, limit := c.rateLimiters.GetAvailableKey(c.apiKeys); key != "" {
			return key, limit, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", 0, errors.New("all API keys are currently rate limited")
}

func (c *Client) getNextClient() (*utils.ClientWrapper, error) {
	startIdx := int(c.proxyIndex.Load())
	maxTries := len(c.clients)

	for i := 0; i < maxTries; i++ {
		idx := (startIdx + i) % len(c.clients)
		wrapper := &c.clients[idx]

		if wrapper.RateLimiter != nil && !wrapper.RateLimiter.Allow() {
			continue
		}

		c.proxyIndex.Store(int32(idx))
		return wrapper, nil
	}

	// If all clients are rate limited, try again after a short delay
	time.Sleep(50 * time.Millisecond)
	idx := int(c.proxyIndex.Add(1)) % len(c.clients)
	return &c.clients[idx], nil
}

func (c *Client) ProxyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.logger.Infof("Received request: %s %s", r.Method, r.URL.Path)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		apiKey, rateLimit, err := c.getNextAPIKey()
		if err != nil {
			c.logger.Error(err)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		limiter := c.rateLimiters.GetLimiter(apiKey, rateLimit)
		if !limiter.Allow() {
			c.logger.Warnf("Rate limit exceeded for key: %s", apiKey)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		clientWrapper, err := c.getNextClient()
		if err != nil {
			c.logger.Error(err)
			http.Error(w, "No available clients", http.StatusServiceUnavailable)
			return
		}

		targetURL, err := url.Parse(c.config.Domain)
		if err != nil {
			c.logger.Errorf("Failed to parse target URL: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		targetURL.Path = r.URL.Path
		targetURL.RawQuery = r.URL.RawQuery

		outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL.String(), r.Body)
		if err != nil {
			c.logger.Errorf("Failed to create request: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for key, values := range r.Header {
			if key != c.config.Key {
				for _, value := range values {
					outReq.Header.Add(key, value)
				}
			}
		}

		outReq.Header.Set(c.config.Key, apiKey)
		outReq.Header.Set("Accept", "application/json")
		outReq.Header.Set("Content-Type", "application/json")

		clientType := "proxy"
		if clientWrapper.IsDirect {
			clientType = "direct"
		}
		c.logger.Debugf("Forwarding request via %s access to %s with API key: %s",
			clientType, targetURL.String(), apiKey)

		resp, err := clientWrapper.Client.Do(outReq)
		if err != nil {
			c.logger.Errorf("Failed to forward request: %v", err)
			http.Error(w, "Failed to forward request", http.StatusBadGateway)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				c.logger.Errorf("Failed to close response body: %v", err)
			}
		}(resp.Body)

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(resp.StatusCode)

		if _, err := io.Copy(w, resp.Body); err != nil {
			c.logger.Errorf("Failed to copy response: %v", err)
			return
		}

		c.logger.Debugf("Request completed with status: %d via %s access",
			resp.StatusCode, clientType)
	})
}

package utils

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

func CreateProxyClient(proxyURL string) (*http.Client, error) {
	// Parse the proxy URL
	urlParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	// Create dialer based on proxy type
	var dialer proxy.Dialer

	switch urlParsed.Scheme {
	case "socks5", "socks5h":
		auth := &proxy.Auth{
			User: urlParsed.User.Username(),
		}
		if password, ok := urlParsed.User.Password(); ok {
			auth.Password = password
		}

		dialSocksProxy, err := proxy.SOCKS5("tcp", urlParsed.Host, auth,
			&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			},
		)
		if err != nil {
			return nil, err
		}
		dialer = dialSocksProxy

	case "http", "https":
		// For HTTP proxies, we can use the proxy URL directly
		proxyURL := func(_ *http.Request) (*url.URL, error) {
			return urlParsed, nil
		}
		transport := &http.Transport{
			Proxy: proxyURL,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		return &http.Client{Transport: transport}, nil
	}

	// Create transport using the dialer
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{Transport: transport}, nil
}

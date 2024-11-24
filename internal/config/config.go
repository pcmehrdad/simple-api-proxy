package config

import (
	"encoding/json"
	"errors"
	"os"
)

type KeyLimit map[string]int

type Config struct {
	Domain          string     `json:"domain"`
	ProxyType       string     `json:"PROXY_TYPE"`
	Key             string     `json:"KEY"`
	Values          []KeyLimit `json:"VALUES"`
	DirectAccess    bool       `json:"DIRECT_ACCESS"`
	DirectAccessTPS int        `json:"DIRECT_ACCESS_TPS"`
	ProxyTPS        int        `json:"PROXY_TPS,omitempty"`
	Proxies         []string   `json:"PROXIES"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Validate required fields
	if config.Domain == "" {
		return nil, errors.New("domain is required")
	}
	if config.Key == "" {
		return nil, errors.New("KEY is required")
	}
	if len(config.Values) == 0 {
		return nil, errors.New("VALUES is required and must not be empty")
	}
	if config.DirectAccess && config.DirectAccessTPS <= 0 {
		return nil, errors.New("DIRECT_ACCESS_TPS must be positive when DIRECT_ACCESS is true")
	}
	if len(config.Proxies) > 0 && config.ProxyTPS < 0 {
		return nil, errors.New("PROXY_TPS cannot be negative")
	}

	return &config, nil
}

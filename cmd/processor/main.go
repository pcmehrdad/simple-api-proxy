package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"api-proxy/internal/api"
	"api-proxy/internal/config"
	"github.com/sirupsen/logrus"
)

func getLogLevel() logrus.Level {
	// First try environment variable
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "info" // default level
	}

	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return logrus.InfoLevel // fallback to info level if parsing fails
	}
	return level
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to config file")
	logLevel := flag.String("log-level", "", "Logging level (debug, info, warn, error)")
	flag.Parse()

	// Set log level from command line flag or environment variable
	if *logLevel != "" {
		level, err := logrus.ParseLevel(*logLevel)
		if err != nil {
			logger.Fatalf("Invalid log level: %v", err)
		}
		logger.SetLevel(level)
	} else {
		logger.SetLevel(getLogLevel())
	}

	// Optional: Add log formatting
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		// You can uncomment the following to show log colors in console
		// ForceColors:   true,
	})

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize API client
	client := api.NewClient(cfg, logger)

	// Ensure port format
	addr := "0.0.0.0:3003"

	// Create server
	server := &http.Server{
		Addr:    addr,
		Handler: client.ProxyHandler(),
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down gracefully...")
		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Server shutdown error: %v", err)
		}
		cancel()
	}()

	// Start server
	logger.WithFields(logrus.Fields{
		"address":   addr,
		"log_level": logger.GetLevel().String(),
	}).Info("Starting server")

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("Server error: %v", err)
	}

	logger.Info("Server stopped gracefully")
}

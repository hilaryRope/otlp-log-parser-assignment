package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"otlp-log-parser-assignment/config"
	"otlp-log-parser-assignment/internal/logger"
	"otlp-log-parser-assignment/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	appLogger, err := logger.New(cfg.Debug)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	startMetricsServer(cfg, appLogger)

	srv, err := server.NewServer(cfg, appLogger)
	if err != nil {
		appLogger.Fatalw("Failed to create server", "error", err)
	}

	if err := srv.Start(); err != nil {
		appLogger.Fatalw("Server error", "error", err)
	}
}

func startMetricsServer(cfg *config.Config, logger *logger.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", cfg.MetricsPort)
	logger.Infow("Starting Prometheus metrics server", "address", addr)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.Errorw("Metrics server failed", "error", err)
		}
	}()
}

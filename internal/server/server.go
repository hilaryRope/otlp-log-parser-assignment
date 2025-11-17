package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	collectorpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"otlp-log-parser-assignment/config"
	"otlp-log-parser-assignment/internal/attributes"
	"otlp-log-parser-assignment/internal/counter"
	"otlp-log-parser-assignment/internal/logger"
	"otlp-log-parser-assignment/internal/service"
)

// Server represents the gRPC server
type Server struct {
	config        *config.Config
	grpcServer    *grpc.Server
	logsService   *service.LogsService
	windowCounter *counter.WindowCounter
	listener      net.Listener
	logger        *logger.Logger
}

func NewServer(cfg *config.Config, logger *logger.Logger) (*Server, error) {
	// Create attribute extractor
	extractor := attributes.NewExtractor(cfg.AttributeKey)

	// Create window counter
	windowCounter := counter.NewWindowCounter(cfg.WindowDuration, logger, cfg.Debug)

	// Create logs service
	logsService := service.NewLogsService(extractor, windowCounter, logger)

	// Create gRPC server with options for high throughput
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(16*1024*1024), // 16MB max message size
		grpc.MaxConcurrentStreams(1000),   // Support many concurrent streams
	)

	// Register services
	collectorpb.RegisterLogsServiceServer(grpcServer, logsService)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for debugging
	reflection.Register(grpcServer)

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return &Server{
		config:        cfg,
		grpcServer:    grpcServer,
		logsService:   logsService,
		windowCounter: windowCounter,
		listener:      listener,
		logger:        logger,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Infow("Starting server",
		"port", s.config.GRPCPort,
		"attribute_key", s.config.AttributeKey,
		"window_duration", s.config.WindowDuration,
		"debug", s.config.Debug,
	)

	// Start window counter
	s.windowCounter.Start()

	// Start gRPC server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		s.logger.Infow("gRPC server listening", "address", s.listener.Addr().String())
		if err := s.grpcServer.Serve(s.listener); err != nil {
			s.logger.Errorw("gRPC server failed", "error", err)
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		s.logger.Infow("Received signal, initiating graceful shutdown", "signal", sig.String())
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.logger.Infow("Shutting down server")

	// Stop accepting new requests
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case <-stopped:
		s.logger.Infow("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.logger.Errorw("Shutdown timeout exceeded, forcing stop")
		s.grpcServer.Stop()
	}

	// Stop window counter
	s.windowCounter.Stop()

	s.logger.Infow("Server shutdown complete")
	return nil
}

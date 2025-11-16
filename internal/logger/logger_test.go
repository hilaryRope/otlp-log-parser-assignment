package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNew_ProductionMode(t *testing.T) {
	logger, err := New(false) // production mode
	if err != nil {
		t.Fatalf("Failed to create production logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be non-nil")
	}

	// Verify it's a sugared logger
	if logger.SugaredLogger == nil {
		t.Fatal("Expected SugaredLogger to be non-nil")
	}
}

func TestNew_DebugMode(t *testing.T) {
	logger, err := New(true) // debug mode
	if err != nil {
		t.Fatalf("Failed to create debug logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be non-nil")
	}

	// Verify it's a sugared logger
	if logger.SugaredLogger == nil {
		t.Fatal("Expected SugaredLogger to be non-nil")
	}
}

func TestWith(t *testing.T) {
	// Create an observed logger to capture log output
	core, recorded := observer.New(zapcore.InfoLevel)
	baseLogger := zap.New(core).Sugar()
	logger := &Logger{baseLogger}

	// Create a logger with context
	contextLogger := logger.With("service", "test-service", "version", "1.0.0")

	// Log a message
	contextLogger.Info("test message")

	// Verify the log was recorded
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	if log.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", log.Message)
	}

	// Verify context fields are present
	fields := log.Context
	serviceField := findField(fields, "service")
	if serviceField == nil || serviceField.String != "test-service" {
		t.Errorf("Expected service field to be 'test-service'")
	}

	versionField := findField(fields, "version")
	if versionField == nil || versionField.String != "1.0.0" {
		t.Errorf("Expected version field to be '1.0.0'")
	}
}

func TestLoggingLevels(t *testing.T) {
	// Create an observed logger to capture all levels
	core, recorded := observer.New(zapcore.DebugLevel)
	baseLogger := zap.New(core).Sugar()
	logger := &Logger{baseLogger}

	// Test different log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Verify all logs were recorded
	logs := recorded.All()
	if len(logs) != 4 {
		t.Fatalf("Expected 4 log entries, got %d", len(logs))
	}

	expectedLevels := []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
	}

	expectedMessages := []string{
		"debug message",
		"info message",
		"warn message",
		"error message",
	}

	for i, log := range logs {
		if log.Level != expectedLevels[i] {
			t.Errorf("Expected level %v, got %v", expectedLevels[i], log.Level)
		}
		if log.Message != expectedMessages[i] {
			t.Errorf("Expected message '%s', got '%s'", expectedMessages[i], log.Message)
		}
	}
}

func TestStructuredLogging(t *testing.T) {
	// Create an observed logger
	core, recorded := observer.New(zapcore.InfoLevel)
	baseLogger := zap.New(core).Sugar()
	logger := &Logger{baseLogger}

	// Log with structured fields
	logger.Infow("processing request",
		"request_id", "req-123",
		"user_id", 456,
		"duration_ms", 250.5,
		"success", true,
	)

	// Verify the log was recorded with fields
	logs := recorded.All()
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	log := logs[0]
	fields := log.Context

	// Check each field
	if field := findField(fields, "request_id"); field == nil || field.String != "req-123" {
		t.Errorf("Expected request_id to be 'req-123'")
	}

	if field := findField(fields, "user_id"); field == nil || field.Integer != 456 {
		t.Errorf("Expected user_id to be 456")
	}

	if field := findField(fields, "duration_ms"); field == nil {
		t.Errorf("Expected duration_ms field to exist")
	}

	if field := findField(fields, "success"); field == nil {
		t.Errorf("Expected success field to exist")
	}
}

func TestProductionJSONOutput(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a production config that writes to our buffer
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}

	// Build logger with custom output
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)

	baseLogger := zap.New(core).Sugar()
	logger := &Logger{baseLogger}

	// Log a structured message
	logger.Infow("test message", "key", "value", "number", 42)

	// Sync to ensure output is written
	_ = logger.Sync() // Ignore error in test

	// Verify JSON output
	output := buf.String()
	if output == "" {
		t.Fatal("Expected JSON output, got empty string")
	}

	// Parse the JSON to verify it's valid
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify expected fields
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg to be 'test message', got %v", logEntry["msg"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("Expected key to be 'value', got %v", logEntry["key"])
	}

	if logEntry["number"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected number to be 42, got %v", logEntry["number"])
	}
}

// Helper function to find a field in zap context
func findField(fields []zapcore.Field, key string) *zapcore.Field {
	for _, field := range fields {
		if field.Key == key {
			return &field
		}
	}
	return nil
}

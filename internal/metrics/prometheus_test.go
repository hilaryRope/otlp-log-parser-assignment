package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRequestsTotal(t *testing.T) {
	// Reset the counter to ensure clean test state
	RequestsTotal.Add(0)
	initialValue := testutil.ToFloat64(RequestsTotal)

	// Increment the counter
	RequestsTotal.Inc()
	RequestsTotal.Inc()

	// Verify the counter increased by 2
	finalValue := testutil.ToFloat64(RequestsTotal)
	expected := initialValue + 2
	if finalValue != expected {
		t.Errorf("Expected RequestsTotal to be %f, got %f", expected, finalValue)
	}
}

func TestLogRecordsProcessed(t *testing.T) {
	// Reset and get initial value
	LogRecordsProcessed.Add(0)
	initialValue := testutil.ToFloat64(LogRecordsProcessed)

	// Add some log records
	LogRecordsProcessed.Add(10)
	LogRecordsProcessed.Add(5)

	// Verify the counter increased by 15
	finalValue := testutil.ToFloat64(LogRecordsProcessed)
	expected := initialValue + 15
	if finalValue != expected {
		t.Errorf("Expected LogRecordsProcessed to be %f, got %f", expected, finalValue)
	}
}

func TestAttributeValuesTotal(t *testing.T) {
	// Test with different attribute values
	testValues := []string{"service-a", "service-b", "service-a"}

	// Record the values
	for _, value := range testValues {
		AttributeValuesTotal.WithLabelValues(value).Inc()
	}

	// Verify service-a has count of 2
	serviceACount := testutil.ToFloat64(AttributeValuesTotal.WithLabelValues("service-a"))
	if serviceACount < 2 {
		t.Errorf("Expected service-a count to be at least 2, got %f", serviceACount)
	}

	// Verify service-b has count of at least 1
	serviceBCount := testutil.ToFloat64(AttributeValuesTotal.WithLabelValues("service-b"))
	if serviceBCount < 1 {
		t.Errorf("Expected service-b count to be at least 1, got %f", serviceBCount)
	}
}

func TestMetricsRegistration(t *testing.T) {
	// Verify that our metrics are properly registered by checking
	// if they appear in the default registry
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	expectedMetrics := []string{
		"otlp_log_parser_assignment_requests_total",
		"otlp_log_parser_assignment_log_records_processed_total",
		"otlp_log_parser_assignment_attribute_values_total",
	}

	foundMetrics := make(map[string]bool)
	for _, mf := range metricFamilies {
		foundMetrics[mf.GetName()] = true
	}

	for _, expected := range expectedMetrics {
		if !foundMetrics[expected] {
			t.Errorf("Expected metric %s not found in registry", expected)
		}
	}
}

func TestMetricsOutput(t *testing.T) {
	// Increment some metrics
	RequestsTotal.Inc()
	LogRecordsProcessed.Add(42)
	AttributeValuesTotal.WithLabelValues("test-service").Inc()

	// Gather metrics and verify they contain expected content
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Convert to string format (similar to what /metrics endpoint would return)
	output := ""
	for _, mf := range metricFamilies {
		if strings.HasPrefix(mf.GetName(), "otlp_log_parser_assignment_") {
			output += mf.String()
		}
	}

	// Verify output contains our metrics
	expectedStrings := []string{
		"otlp_log_parser_assignment_requests_total",
		"otlp_log_parser_assignment_log_records_processed_total",
		"otlp_log_parser_assignment_attribute_values_total",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected metrics output to contain %s", expected)
		}
	}
}

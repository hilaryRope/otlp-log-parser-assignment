package service

import (
	"context"
	"testing"
	"time"

	collectorpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"otlp-log-parser-assignment/internal/attributes"
	"otlp-log-parser-assignment/internal/counter"
	"otlp-log-parser-assignment/internal/logger"
)

func TestLogsService_Export_NilRequest(t *testing.T) {
	extractor := attributes.NewExtractor("test.key")
	testLogger, _ := logger.New(false)
	wc := counter.NewWindowCounter(1*time.Second, testLogger, false)
	svc := NewLogsService(extractor, wc, testLogger)

	resp, err := svc.Export(context.Background(), nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Error("Expected non-nil response")
	}
}

func TestLogsService_Export_EmptyRequest(t *testing.T) {
	extractor := attributes.NewExtractor("test.key")
	testLogger, _ := logger.New(false)
	wc := counter.NewWindowCounter(1*time.Second, testLogger, false)
	svc := NewLogsService(extractor, wc, testLogger)

	req := &collectorpb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{},
	}

	resp, err := svc.Export(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp == nil {
		t.Error("Expected non-nil response")
	}

	counts := wc.GetCurrentCounts()
	if len(counts) != 0 {
		t.Errorf("Expected no counts, got %v", counts)
	}
}

func TestLogsService_Export_LogLevelAttribute(t *testing.T) {
	extractor := attributes.NewExtractor("foo")
	testLogger, _ := logger.New(false)
	wc := counter.NewWindowCounter(1*time.Second, testLogger, false)
	svc := NewLogsService(extractor, wc, testLogger)

	req := &collectorpb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				ScopeLogs: []*logspb.ScopeLogs{
					{
						LogRecords: []*logspb.LogRecord{
							{
								Attributes: []*commonpb.KeyValue{
									{
										Key: "foo",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{StringValue: "bar"},
										},
									},
								},
							},
							{
								Attributes: []*commonpb.KeyValue{
									{
										Key: "foo",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{StringValue: "baz"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := svc.Export(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	counts := wc.GetCurrentCounts()
	if counts["bar"] != 1 {
		t.Errorf("Expected count for 'bar' to be 1, got %d", counts["bar"])
	}
	if counts["baz"] != 1 {
		t.Errorf("Expected count for 'baz' to be 1, got %d", counts["baz"])
	}
}

func TestLogsService_Export_ResourceLevelAttribute(t *testing.T) {
	extractor := attributes.NewExtractor("service.name")
	testLogger, _ := logger.New(false)
	wc := counter.NewWindowCounter(1*time.Second, testLogger, false)
	svc := NewLogsService(extractor, wc, testLogger)

	req := &collectorpb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "my-service"},
							},
						},
					},
				},
				ScopeLogs: []*logspb.ScopeLogs{
					{
						LogRecords: []*logspb.LogRecord{
							{}, // No log-level attribute
							{}, // No log-level attribute
						},
					},
				},
			},
		},
	}

	_, err := svc.Export(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	counts := wc.GetCurrentCounts()
	if counts["my-service"] != 2 {
		t.Errorf("Expected count for 'my-service' to be 2, got %d", counts["my-service"])
	}
}

func TestLogsService_Export_AttributePriority(t *testing.T) {
	extractor := attributes.NewExtractor("env")
	testLogger, _ := logger.New(false)
	wc := counter.NewWindowCounter(1*time.Second, testLogger, false)
	svc := NewLogsService(extractor, wc, testLogger)

	req := &collectorpb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key: "env",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{StringValue: "resource-level"},
							},
						},
					},
				},
				ScopeLogs: []*logspb.ScopeLogs{
					{
						LogRecords: []*logspb.LogRecord{
							{
								// Log-level attribute should take priority
								Attributes: []*commonpb.KeyValue{
									{
										Key: "env",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{StringValue: "log-level"},
										},
									},
								},
							},
							{
								// Should fall back to resource-level
							},
						},
					},
				},
			},
		},
	}

	_, err := svc.Export(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	counts := wc.GetCurrentCounts()
	if counts["log-level"] != 1 {
		t.Errorf("Expected count for 'log-level' to be 1, got %d", counts["log-level"])
	}
	if counts["resource-level"] != 1 {
		t.Errorf("Expected count for 'resource-level' to be 1, got %d", counts["resource-level"])
	}
}

func TestLogsService_Export_UnknownAttribute(t *testing.T) {
	extractor := attributes.NewExtractor("missing.key")
	testLogger, _ := logger.New(false)
	wc := counter.NewWindowCounter(1*time.Second, testLogger, false)
	svc := NewLogsService(extractor, wc, testLogger)

	req := &collectorpb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				ScopeLogs: []*logspb.ScopeLogs{
					{
						LogRecords: []*logspb.LogRecord{
							{
								Attributes: []*commonpb.KeyValue{
									{
										Key: "other.key",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{StringValue: "value"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := svc.Export(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	counts := wc.GetCurrentCounts()
	if counts[attributes.UnknownValue] != 1 {
		t.Errorf("Expected count for 'unknown' to be 1, got %d", counts[attributes.UnknownValue])
	}
}

func TestLogsService_countLogRecords(t *testing.T) {
	svc := &LogsService{}

	tests := []struct {
		name         string
		resourceLogs []*logspb.ResourceLogs
		want         int
	}{
		{
			name:         "nil input",
			resourceLogs: nil,
			want:         0,
		},
		{
			name:         "empty input",
			resourceLogs: []*logspb.ResourceLogs{},
			want:         0,
		},
		{
			name: "single log record",
			resourceLogs: []*logspb.ResourceLogs{
				{
					ScopeLogs: []*logspb.ScopeLogs{
						{
							LogRecords: []*logspb.LogRecord{{}},
						},
					},
				},
			},
			want: 1,
		},
		{
			name: "multiple log records",
			resourceLogs: []*logspb.ResourceLogs{
				{
					ScopeLogs: []*logspb.ScopeLogs{
						{
							LogRecords: []*logspb.LogRecord{{}, {}, {}},
						},
						{
							LogRecords: []*logspb.LogRecord{{}, {}},
						},
					},
				},
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.countLogRecords(tt.resourceLogs)
			if got != tt.want {
				t.Errorf("countLogRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

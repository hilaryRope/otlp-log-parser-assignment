package service

import (
	"context"

	collectorpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	"otlp-log-parser-assignment/internal/attributes"
	"otlp-log-parser-assignment/internal/counter"
	"otlp-log-parser-assignment/internal/logger"
	"otlp-log-parser-assignment/internal/metrics"
)

type LogsService struct {
	collectorpb.UnimplementedLogsServiceServer
	extractor *attributes.Extractor
	counter   *counter.WindowCounter
	logger    *logger.Logger
}

func NewLogsService(extractor *attributes.Extractor, counter *counter.WindowCounter, logger *logger.Logger) *LogsService {
	return &LogsService{
		extractor: extractor,
		counter:   counter,
		logger:    logger.With("component", "service"),
	}
}

func (s *LogsService) Export(ctx context.Context, req *collectorpb.ExportLogsServiceRequest) (*collectorpb.ExportLogsServiceResponse, error) {
	if req == nil {
		s.logger.Infow("Received nil request")
		return &collectorpb.ExportLogsServiceResponse{}, nil
	}

	// Count log records for metrics
	logRecordCount := s.countLogRecords(req.ResourceLogs)

	// Process logs in batch for high throughput using hierarchy extraction
	attributeValues := s.extractAttributeValues(req.ResourceLogs)

	s.logger.Infow("Processing request", "log_records", logRecordCount, "attribute_values", len(attributeValues))

	// Record metrics
	metrics.RequestsTotal.Inc()
	metrics.LogRecordsProcessed.Add(float64(logRecordCount))
	for _, value := range attributeValues {
		metrics.AttributeValuesTotal.WithLabelValues(value).Inc()
	}

	s.counter.IncrementBatch(attributeValues)

	// Return success response
	// Note: OTLP supports PartialSuccess for reporting non-fatal errors
	return &collectorpb.ExportLogsServiceResponse{
		PartialSuccess: &collectorpb.ExportLogsPartialSuccess{
			RejectedLogRecords: 0,
			ErrorMessage:       "",
		},
	}, nil
}

// extractAttributeValues extracts all attribute values from the log request
func (s *LogsService) extractAttributeValues(resourceLogs []*logspb.ResourceLogs) []string {
	var values []string

	for _, resourceLog := range resourceLogs {
		if resourceLog == nil {
			continue
		}

		for _, scopeLog := range resourceLog.ScopeLogs {
			if scopeLog == nil {
				continue
			}

			for _, logRecord := range scopeLog.LogRecords {
				if logRecord == nil {
					continue
				}

				// Use hierarchy extraction method like the starter code
				value := s.extractor.ExtractFromHierarchy(resourceLog.Resource, scopeLog.Scope, logRecord)
				values = append(values, value)
			}
		}
	}

	return values
}

// countLogRecords counts the total number of log records in the request
func (s *LogsService) countLogRecords(resourceLogs []*logspb.ResourceLogs) int {
	count := 0
	for _, resourceLog := range resourceLogs {
		if resourceLog == nil {
			continue
		}
		for _, scopeLog := range resourceLog.ScopeLogs {
			if scopeLog == nil {
				continue
			}
			count += len(scopeLog.LogRecords)
		}
	}
	return count
}

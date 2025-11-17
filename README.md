# OTLP Log Parser Assignment

A gRPC service that receives OpenTelemetry Protocol (OTLP) log records and aggregates them based on configurable attributes. 
The service tracks log records per distinct attribute value and reports counts within configurable time windows.

## Assignment Requirements

This implementation fulfills the following requirements:
- ✅ Receives log records on a gRPC endpoint (OTLP protocol)
- ✅ Configurable attribute key and duration
- ✅ Counts unique log records per distinct attribute value
- ✅ Prints delta counts to stdout within each window
- ✅ Searches attributes at `Resource`, `Scope`, and `Log` levels
- ✅ Handles throughput (messages and records per message)

## Features

- **OTLP Compliant**: OpenTelemetry Protocol specification compliance
- **Multi-Level Attribute Extraction**: Searches attributes at `Resource`, `Scope`, and `Log` levels (priority: Log > Scope > Resource)
- **Windowed Aggregation**: Configurable time windows with formatted output and statistics
- **High Throughput**: Optimised for large volumes with batch operations and concurrent processing
- **Structured Logging**: JSON-formatted logs using `zap` for production observability
- **Prometheus Metrics**: Exposes `/metrics` endpoint for monitoring and alerting
- **Health Checks**: Built-in gRPC health check service
- **Graceful Shutdown**: Proper signal handling and resource cleanup
- **Configuration Validation**: Comprehensive input validation with helpful error messages

## Quick Start

### Using Docker Compose (Recommended)
```bash
docker compose up -d
docker compose logs -f
```

### Using Make - Building the binary
```bash
make build
./otlp-log-parser-assignment -attribute-key=foo -window-duration=10s
```

## How It Works

**OTLP Request** → **Extract Attributes** (Resource→Scope→Log priority) → **Count by Value** → **Report Every Window**

The service processes OTLP log records, extracts a configurable attribute from three levels (Resource, Scope, or Log), counts occurrences per unique value, and reports aggregated counts at each time window with enhanced formatting and statistics.

## Installation & Usage

### Prerequisites
- Go 1.23
- Docker & Docker Compose (for containerized deployment)
- Make (optional)

### Running the Server

**Option 1: Docker Compose (Recommended)**
```bash
# Production mode (JSON logs only)
docker compose up -d
docker compose logs -f

# Debug mode (JSON logs + ASCII tables)
DEBUG=true docker compose up -d
docker compose logs -f
```

**Option 2: Docker CLI**
```bash
docker build -t otlp-log-parser-assignment .
docker run -d -p 4317:4317 -p 9090:9090 \
  -e ATTRIBUTE_KEY=foo \
  -e WINDOW_DURATION=30s \
  -e DEBUG=true \
  otlp-log-parser-assignment
```

**Option 3: Build from Source**
```bash
# Build
make build
# or
go build -o otlp-log-parser-assignment ./cmd

# Run with custom configuration
./otlp-log-parser-assignment -attribute-key=foo -window-duration=10s -debug=true
```

**Option 4: Run Directly (Development)**
```bash
go run ./cmd -attribute-key=foo -window-duration=10s -debug=true
```

### Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `4317` | gRPC server port |
| `-metrics-port` | `9090` | Port for Prometheus metrics endpoint |
| `-attribute-key` | `service.name` | Attribute key to track across Resource/Scope/Log levels |
| `-window-duration` | `10s` | Time window for aggregating and reporting counts |
| `-debug` | `false` | Enable debug mode: JSON logs + ASCII tables with percentages |


## Testing

### Run All Tests

```bash
go test ./...

# Or with Make
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

This generates a `coverage.html` file that you can open in your browser.

## Testing the API

### Using grpcurl

To test the service, you can use `grpcurl`, a command-line tool for interacting with gRPC servers.

**Installation (macOS):**
```bash
brew install grpcurl
```

**Send a Test Log:**

```bash
grpcurl -plaintext -d @ localhost:4317 \
  opentelemetry.proto.collector.logs.v1.LogsService/Export <<EOF
{
  "resourceLogs": [{
    "scopeLogs": [{
      "logRecords": [{
        "body": {"stringValue": "test log"},
        "attributes": [{
          "key": "foo",
          "value": {"stringValue": "bar"}
        }]
      }]
    }]
  }]
}
EOF
```

### Using the API Testing Guide

For additional testing scenarios and examples, see the [API Testing Guide](api-testing/README.md).

### Test Structure

- `config/config_test.go` - Configuration validation tests
- `internal/attributes/extractor_test.go` - Attribute extraction logic tests
- `internal/counter/window_counter_test.go` - Window counter and aggregation tests
- `internal/service/logs_service_test.go` - OTLP service handler tests


## Implementation Details


### Design for High Throughput

The implementation includes several design decisions to support high throughput:

- **Batch Processing**: Attribute values are extracted in bulk and incremented in a single operation to minimize lock contention
- **Thread-Safe Counters**: Uses `sync.RWMutex` for concurrent access to shared state
- **gRPC Configuration**: Configured with 16MB max message size and support for 1000 concurrent streams

### Graceful Shutdown

The server handles `SIGINT` and `SIGTERM` signals gracefully:
1. Stops accepting new requests
2. Waits for in-flight requests to complete (30s timeout)
3. Reports final window counts
4. Cleans up resources

### Observability

**Enhanced Structured Logging** (using `zap`):
- JSON-formatted logs with detailed metrics
- **Debug Mode** (`-debug=true`): JSON logs + beautiful ASCII tables
- **Window Reports** with comprehensive statistics:
  - Individual attribute counts with percentages
  - Time ranges, duration, total logs, unique values
  - Example: `"attribute_counts":{"service-name":{"count":42,"percentage":85.7}}`
- **Request Processing**: Structured fields for log records and attribute values
- **Component-scoped**: server, service, counter components
- **Proper log levels**: DEBUG, INFO, WARN, ERROR

**Prometheus Metrics** (available at `http://localhost:9090/metrics`):
- `otlp_log_parser_assignment_requests_total` - Total number of requests received
- `otlp_log_parser_assignment_log_records_processed_total` - Total log records processed
- `otlp_log_parser_assignment_attribute_values_total` - Count by attribute value (with labels)

**Health Checks**:
- gRPC health check service available
- Returns `SERVING` status when operational
- Can be queried using gRPC health check protocol

**Reflection**:
- gRPC reflection enabled for debugging with tools like `grpcurl`

## Architecture

### System Overview

```
OTLP Client → gRPC Server (Port 4317) → LogsService
                     ↓                        ↓
            Prometheus Metrics         ┌─────────────┴──────────────┐
            (Port 9090)                ↓                            ↓
                     ↓           AttributeExtractor              WindowCounter
              /metrics endpoint   (Resource→Scope→Log)            (Aggregation)
                                        ↓                            ↓
                                   Extract Values              Count by Value
                                        ↓                            ↓
                                 Prometheus Counters         Structured Logs
```

### Components

- **LogsService** - Handles OTLP gRPC requests, orchestrates processing with structured logging
- **AttributeExtractor** - Extracts attribute values with priority: Log > Scope > Resource
- **WindowCounter** - Thread-safe aggregation with configurable time windows and structured reporting
- **Prometheus Metrics** - Exposes counters for requests, log records, and attribute values
- **Structured Logger** - Zap-based JSON logging for production observability
- **Server** - gRPC server with health checks, graceful shutdown, and metrics endpoint

### Data Flow

1. **Receive** OTLP log request via gRPC (with structured logging)
2. **Validate** request and count log records
3. **Extract** attribute values from Resource/Scope/Log levels (batch operation)
4. **Record** Prometheus metrics (requests, log records, attribute values)
5. **Increment** window counters (thread-safe batch update)
6. **Report** aggregated counts every window with structured JSON logs
7. **Return** OTLP PartialSuccess response
8. **Expose** metrics via `/metrics` endpoint for monitoring systems


### Project Structure

```
├── api-testing/             # API testing with grpcurl and curl
├── cmd/                      # Main application entry point
├── config/                   # Configuration management with validation
├── internal/
│   ├── attributes/          # Attribute extraction logic
│   ├── counter/             # Window-based counting with structured logging
│   ├── logger/              # Zap-based structured logging
│   ├── metrics/             # Prometheus metrics definitions and tests
│   ├── service/             # OTLP logs service with observability
│   └── server/              # gRPC server with health checks and metrics
├── vendor/                  # Vendored dependencies
├── .gitignore               # Git ignore file
├── Dockerfile               # Docker deployment
├── docker-compose.yml       # Docker Compose configuration
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── Makefile                 # Build automation
└── README.md                # Project documentation
```

## Monitoring & Metrics

### Prometheus Metrics

The service exposes metrics at `http://localhost:9090/metrics` in Prometheus format:

```bash
# Check if metrics endpoint is available
curl http://localhost:9090/metrics

# Example metrics output:
# HELP otlp_log_parser_assignment_requests_total Total number of OTLP log export requests received.
# TYPE otlp_log_parser_assignment_requests_total counter
otlp_log_parser_assignment_requests_total 42

# HELP otlp_log_parser_assignment_log_records_processed_total Total number of log records processed.
# TYPE otlp_log_parser_assignment_log_records_processed_total counter
otlp_log_parser_assignment_log_records_processed_total 1337

# HELP otlp_log_parser_assignment_attribute_values_total Total number of times each attribute value has been seen.
# TYPE otlp_log_parser_assignment_attribute_values_total counter
otlp_log_parser_assignment_attribute_values_total{value="service-a"} 500
otlp_log_parser_assignment_attribute_values_total{value="service-b"} 837
```

### Health Check

```bash
# Check service health using grpcurl
grpcurl -plaintext localhost:4317 grpc.health.v1.Health/Check
```

## Example Output

### Default Structured JSON
Enhanced structured JSON logs with detailed statistics:

```json
2025-11-17T19:45:57.812+0100    INFO    service/logs_service.go:41      Processing request      {"component": "service", "log_records": 5, "attribute_values": 5}
2025-11-17T19:45:58.427+0100    INFO    counter/window_counter.go:126   Log attribute counts report     {"component": "counter", "window_number": 2, "time_range": "19:44:08 - 19:45:58", "duration": "1m50s", "total_logs": 5, "unique_values": 4, "attribute_counts": {"bar":{"count":1,"percentage":20},"baz":{"count":2,"percentage":40},"qux":{"count":1,"percentage":20},"unknown":{"count":1,"percentage":20}}}
```

### Debug Mode (`-debug=true`)
JSON logs **plus** beautiful ASCII tables for human readability:
```
╔═══════════════════════════════════════════════════════════╗
║          Log Attribute Counts Report                      ║
╠═══════════════════════════════════════════════════════════╣
║ Window #1                                                 ║
║ Time Range: 18:56:44 - 18:57:44                           ║
║ Duration: 1m0s                                            ║
║ Total Logs: 1005                                          ║
║ Unique Values: 3                                          ║
╠═══════════════════════════════════════════════════════════╣
║ Attribute Value Counts:                                   ║
╠═══════════════════════════════════════════════════════════╣
║ bar                                          335 ( 33.3%) ║
║ baz                                          335 ( 33.3%) ║
║ qux                                          335 ( 33.3%) ║
╚═══════════════════════════════════════════════════════════╝
```

### Output Mode Benefits

**Production Mode** (default):
- ✅ **Machine readable** - Perfect for log aggregation systems
- ✅ **Structured fields** - Easy parsing and filtering
- ✅ **Detailed metrics** - Individual counts and percentages in JSON
- ✅ **Performance optimized** - Minimal overhead

**Debug Mode** (`-debug=true`):
- ✅ **Human readable** - ASCII tables for immediate visual analysis
- ✅ **Development friendly** - Easy to read during testing and debugging
- ✅ **Complete information** - Both JSON logs AND visual tables
- ✅ **Best of both worlds** - Machine + human readable simultaneously


## OTLP Protocol Details

### Hierarchy Structure
```
ExportLogsServiceRequest
└── ResourceLogs[]              # Service/host level
    ├── Resource.Attributes[]   # e.g., service.name="my-service"
    └── ScopeLogs[]             # Library/instrumentation level
        ├── Scope.Attributes[]  # e.g., library.version="1.0.0"
        └── LogRecords[]        # Individual logs
            └── Attributes[]    # e.g., http.status_code=200
```

### Attribute Priority
When searching for an attribute key:
1. **Check Log-level first** (highest priority)
2. **Fall back to Scope-level** if not found
3. **Fall back to Resource-level** if still not found
4. **Return "unknown"** if not at any level

### Supported Attribute Types
| Type | Example Input | Output |
|------|---------------|--------|
| String | `"my-service"` | `"my-service"` |
| Integer | `42` | `"42"` |
| Double | `3.14` | `"3.140000"` |
| Boolean | `true` | `"true"` |
| Array | `["prod","critical"]` | `["prod","critical"]` |
| Map | `{"amount":100,"currency":"USD"}` | `{"amount":100,"currency":"USD"}` |
| Bytes | `[72,101,108,108,111]` | `"base64:48656c6c6f"` |

## References

- [OpenTelemetry Logs](https://opentelemetry.io/docs/concepts/signals/logs/)
- [OpenTelemetry Protocol (OTLP)](https://github.com/open-telemetry/opentelemetry-proto)
- [OTLP Logs Examples](https://github.com/open-telemetry/opentelemetry-proto/blob/main/examples/logs.json)

## Assignment Timeline and Time Management
I allocated the four hours as follows:

- Ideation – reviewing requirements, planning the project structure, and outlining the approach: 30 minutes
- Implementation – core development work: 2 hours and 30 minutes
- Testing – API testing (e.g., via grpcurl), running the application, and validating behaviour: 1 hour
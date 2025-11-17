# API Testing

This folder contains testing instructions and collections for the OTLP Log Parser gRPC endpoints and Prometheus metrics using **grpcurl** (recommended) and **curl** for HTTP endpoints.

> 💡 **See the [main README](../README.md) for server installation, configuration options, and architecture details.**

## Quick Start

### 1. Start the Server

```bash
docker compose up -d
docker compose logs -f  # Watch server output
```

> 💡 **Recommended:** Docker Compose ensures a clean, isolated environment. Alternatively, you can [build and run the binary](../README.md#quick-start).

### 2. Install grpcurl (if not already installed)

```bash
# macOS
brew install grpcurl

# Or download from: https://github.com/fullstorydev/grpcurl/releases
```

### 3. Verify Server Reflection

The server has gRPC reflection enabled by default. You can verify available services:

```bash
# List all available services
grpcurl -plaintext localhost:4317 list

# Expected output:
# grpc.health.v1.Health
# grpc.reflection.v1.ServerReflection
# grpc.reflection.v1alpha.ServerReflection
# opentelemetry.proto.collector.logs.v1.LogsService
```

### 4. Send the First Request

#### **Start with HTTP**
```bash
# Test Prometheus metrics endpoint
curl http://localhost:9090/metrics

# Check specific application metrics
curl -s http://localhost:9090/metrics | grep otlp_log_parser_assignment
```

#### **Test gRPC Endpoints**
```bash
# Test health check
grpcurl -plaintext localhost:4317 grpc.health.v1.Health/Check
```

**Expected Health Check Response:**
```json
{
  "status": "SERVING"
}
```

**Expected Server Logs (JSON format):**
```json
{"level":"info","time":"2024-11-16T19:15:30.123Z","component":"service","message":"Processing request","log_records":1,"attribute_values":1}
```

## Available Test Commands

Here are the 6 main test scenarios you can run:

### gRPC Endpoints (Port 4317)

#### 1. **Health Check**
```bash
grpcurl -plaintext localhost:4317 grpc.health.v1.Health/Check
```

#### 2. **Export Logs - Simple**
```bash
grpcurl -plaintext -d '{
  "resourceLogs": [{
    "scopeLogs": [{
      "logRecords": [{
        "body": {"stringValue": "Simple test log message"},
        "attributes": [{
          "key": "service.name",
          "value": {"stringValue": "bar"}
        }]
      }]
    }]
  }]
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```

**Expected OTLP Export Response:**
```json
{
  "partialSuccess": {}
}
```

#### 3. **Export Logs - Complex Types**
```bash
grpcurl -plaintext -d '{
  "resourceLogs": [{
    "scopeLogs": [{
      "logRecords": [{
        "body": {"stringValue": "Complex types test"},
        "attributes": [{
          "key": "tags",
          "value": {
            "arrayValue": {
              "values": [
                {"stringValue": "production"},
                {"stringValue": "critical"}
              ]
            }
          }
        }]
      }]
    }]
  }]
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```

#### 4. **Export Logs - Batch**
```bash
grpcurl -plaintext -d '{
  "resourceLogs": [{
    "scopeLogs": [{
      "logRecords": [
        {
          "body": {"stringValue": "Batch log 1"},
          "attributes": [{
            "key": "service.name",
            "value": {"stringValue": "batch-service-1"}
          }]
        },
        {
          "body": {"stringValue": "Batch log 2"},
          "attributes": [{
            "key": "service.name",
            "value": {"stringValue": "batch-service-2"}
          }]
        }
      ]
    }]
  }]
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```

#### 5. **Export Logs - Attribute Priority Test**
```bash
grpcurl -plaintext -d '{
  "resourceLogs": [{
    "resource": {
      "attributes": [{
        "key": "env",
        "value": {"stringValue": "resource-level"}
      }]
    },
    "scopeLogs": [{
      "scope": {
        "attributes": [{
          "key": "env",
          "value": {"stringValue": "scope-level"}
        }]
      },
      "logRecords": [
        {
          "body": {"stringValue": "Priority test - log level wins"},
          "attributes": [{
            "key": "env",
            "value": {"stringValue": "log-level"}
          }]
        },
        {
          "body": {"stringValue": "Priority test - scope level wins"}
        }
      ]
    }]
  }]
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```

### HTTP Endpoints (Port 9090)

#### 6. **Prometheus Metrics**
```bash
# Get all metrics
curl http://localhost:9090/metrics

# Get only application metrics
curl -s http://localhost:9090/metrics | grep otlp_log_parser_assignment
```

## Test Scenarios

### Basic Test
```bash
# Start with Docker Compose (uses default config)
docker compose up -d
docker compose logs -f

# In Insomnia: Send "Export Logs - Simple"
# Expected: Structured JSON logs showing request processing
# Expected: Window report after 10s with "bar - 1"
```

### Metrics Test
```bash
# Start the server
docker compose up -d

# In Insomnia: Send "Prometheus Metrics"
# Expected: Prometheus metrics in text format
# Example output:
# otlp_log_parser_assignment_requests_total 0
# otlp_log_parser_assignment_log_records_processed_total 0
```

### Complex Types
```bash
# Start with custom attribute key
ATTRIBUTE_KEY=tags docker compose up -d
docker compose logs -f

# In Insomnia: Send "Export Logs - Complex Types"
# Expected: Server shows ["production","critical"] in structured logs
```

### Attribute Priority
```bash
# Start with custom attribute key
ATTRIBUTE_KEY=env docker compose up -d
docker compose logs -f

# In Insomnia: Send "Export Logs - Attribute Priority Test"
# Expected: Structured logs showing both "log-level" and "scope-level" values
```

> **Tip:** Stop the server with `docker compose down` before changing configuration.

## Troubleshooting

**gRPC Connection refused?**
- Check server is running: `docker compose ps` or `lsof -i :4317`
- Start server: `docker compose up -d`

**Metrics endpoint not accessible?**
- Check metrics port: `lsof -i :9090`
- Verify metrics server started: Look for "Starting Prometheus metrics server" in logs
- Try: `curl http://localhost:9090/metrics`

**No proto definitions?**
- Ensure "Use Server Reflection" is enabled in Insomnia
- Server has reflection enabled by default

**No structured logs appearing?**
- Check logs: `docker compose logs -f`
- Logs are in JSON format by default (production mode)
- For human-readable logs, add `-debug=true` flag

**No window reports?**
- Wait for window duration to elapse (default 10s)
- Ensure requests are being sent successfully
- Check for structured log entries with "window_number" field

## Monitoring Results

While testing, monitor the application:

```bash
# Watch server logs (structured JSON)
docker compose logs -f otlp-log-parser-assignment

# Monitor metrics in real-time
watch -n 1 'curl -s http://localhost:9090/metrics | grep otlp_log_parser_assignment'
```

## Collection Files

- **`insomnia-collection.yaml`** - Insomnia collection (optional GUI alternative)

## Resources

- [Main README](../README.md) - Full documentation
- [grpcurl Documentation](https://github.com/fullstorydev/grpcurl)
- [Insomnia gRPC Guide](https://docs.insomnia.rest/insomnia/requests#grpc) (optional)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)

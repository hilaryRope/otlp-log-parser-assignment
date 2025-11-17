# API Testing Guide

Simple examples to test the OTLP Log Parser service.

## Quick Start

### 1. Start the Server
```bash
go run ./cmd -attribute-key=foo -window-duration=10s -debug
```

### 2. Install grpcurl
```bash
brew install grpcurl
```

## Basic Tests

### Health Check
```bash
grpcurl -plaintext localhost:4317 grpc.health.v1.Health/Check
```
**Expected:** `{"status": "SERVING"}`

### Metrics Check
```bash
curl http://localhost:9090/metrics
```
**Expected:** Prometheus metrics output

## Log Testing Examples

### Example 1: Simple Single Log
```bash
grpcurl -plaintext -d '{
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
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```
**Expected:** `{"partialSuccess": {}}`  
**Server Output:**

```bash2025-11-17T19:44:08.426+0100    INFO    counter/window_counter.go:126   Log attribute counts report     {"component": "counter", "window_number": 1, "time_range": "19:42:58 - 19:44:08", "duration": "1m10.001s", "total_logs": 1, "unique_values": 1, "attribute_counts": {"bar":{"count":1,"percentage":100}}}

╔═══════════════════════════════════════════════════════════╗
║          Log Attribute Counts Report                      ║
╠═══════════════════════════════════════════════════════════╣
║ Window #1                                                 ║
║ Time Range: 19:42:58 - 19:44:08                           ║
║ Duration: 1m10.001s                                       ║
║ Total Logs: 1                                             ║
║ Unique Values: 1                                          ║
╠═══════════════════════════════════════════════════════════╣
║ Attribute Value Counts:                                   ║
╠═══════════════════════════════════════════════════════════╣
║ bar                                             1 (100.0%) ║
╚═══════════════════════════════════════════════════════════╝```
```

### Example 2: Sample Attributes
Test the exact pseudo code example from the assignment requirements:
```bash
grpcurl -plaintext -d '{
  "resourceLogs": [{
    "scopeLogs": [{
      "logRecords": [
        {"body": {"stringValue": "my log body 1"}, "attributes": [{"key": "foo", "value": {"stringValue": "bar"}}, {"key": "baz", "value": {"stringValue": "qux"}}]},
        {"body": {"stringValue": "my log body 2"}, "attributes": [{"key": "foo", "value": {"stringValue": "qux"}}, {"key": "baz", "value": {"stringValue": "qux"}}]},
        {"body": {"stringValue": "my log body 3"}, "attributes": [{"key": "baz", "value": {"stringValue": "qux"}}]},
        {"body": {"stringValue": "my log body 4"}, "attributes": [{"key": "foo", "value": {"stringValue": "baz"}}]},
        {"body": {"stringValue": "my log body 5"}, "attributes": [{"key": "foo", "value": {"stringValue": "baz"}}, {"key": "baz", "value": {"stringValue": "qux"}}]}
      ]
    }]
  }]
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```
**Expected:** `{"partialSuccess": {}}`  
**Server Output:**

```bash

2025-11-17T19:45:58.427+0100    INFO    counter/window_counter.go:126   Log attribute counts report     {"component": "counter", "window_number": 2, "time_range": "19:44:08 - 19:45:58", "duration": "1m50s", "total_logs": 5, "unique_values": 4, "attribute_counts": {"bar":{"count":1,"percentage":20},"baz":{"count":2,"percentage":40},"qux":{"count":1,"percentage":20},"unknown":{"count":1,"percentage":20}}}
╔═══════════════════════════════════════════════════════════╗
║          Log Attribute Counts Report                      ║
╠═══════════════════════════════════════════════════════════╣
║ Window #2                                                 ║
║ Time Range: 19:44:08 - 19:45:58                           ║
║ Duration: 1m50s                                           ║
║ Total Logs: 5                                             ║
║ Unique Values: 4                                          ║
╠═══════════════════════════════════════════════════════════╣
║ Attribute Value Counts:                                   ║
╠═══════════════════════════════════════════════════════════╣
║ bar                                             1 ( 20.0%) ║
║ baz                                             2 ( 40.0%) ║
║ qux                                             1 ( 20.0%) ║
║ unknown                                         1 ( 20.0%) ║
╚═══════════════════════════════════════════════════════════╝

```

### Example 3: Multi-Level Attributes
Test attribute priority (Log > Scope > Resource):
```bash
grpcurl -plaintext -d '{
  "resourceLogs": [{
    "resource": {"attributes": [{"key": "foo", "value": {"stringValue": "resource-level"}}]},
    "scopeLogs": [{
      "scope": {"attributes": [{"key": "foo", "value": {"stringValue": "scope-level"}}]},
      "logRecords": [{
        "body": {"stringValue": "priority test"},
        "attributes": [{"key": "foo", "value": {"stringValue": "log-level"}}]
      }]
    }]
  }]
}' localhost:4317 opentelemetry.proto.collector.logs.v1.LogsService/Export
```
**Expected:** `{"partialSuccess": {}}`  
**Server Output:**

```bash

2025-11-17T19:47:28.441+0100    INFO    counter/window_counter.go:126   Log attribute counts report     {"component": "counter", "window_number": 3, "time_range": "19:45:58 - 19:47:28", "duration": "1m30s", "total_logs": 1, "unique_values": 1, "attribute_counts": {"log-level":{"count":1,"percentage":100}}}

╔═══════════════════════════════════════════════════════════╗
║          Log Attribute Counts Report                      ║
╠═══════════════════════════════════════════════════════════╣
║ Window #3                                                 ║
║ Time Range: 19:45:58 - 19:47:28                           ║
║ Duration: 1m30s                                           ║
║ Total Logs: 1                                             ║
║ Unique Values: 1                                          ║
╠═══════════════════════════════════════════════════════════╣
║ Attribute Value Counts:                                   ║
╠═══════════════════════════════════════════════════════════╣
║ log-level                                       1 (100.0%) ║
╚═══════════════════════════════════════════════════════════╝

```


## Troubleshooting

### Port Already in Use
```bash
lsof -ti :4317 | xargs kill -9 2>/dev/null || true
lsof -ti :9090 | xargs kill -9 2>/dev/null || true
```

### Server Not Responding
1. Check if server is running: `ps aux | grep otlp`
2. Check server logs for errors
3. Verify ports with: `lsof -i :4317`

## What to Expect

### Immediate Response
All gRPC calls return: `{"partialSuccess": {}}`

### Server Logs (after window duration)
- **JSON format**: Structured log with counts and percentages
- **Debug mode**: JSON + ASCII table
- **Example**: `{"attribute_counts":{"bar":{"count":1,"percentage":100}}}`

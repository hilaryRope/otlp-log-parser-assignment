package attributes

import (
	"encoding/json"
	"fmt"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

const (
	UnknownValue = "unknown"
)

// Extractor extracts attribute values from OTLP data structures
type Extractor struct {
	attributeKey string
}

func NewExtractor(attributeKey string) *Extractor {
	return &Extractor{
		attributeKey: attributeKey,
	}
}

// ExtractFromHierarchy extracts attribute value with proper hierarchy precedence
// Priority: Log-level > Scope-level > Resource-level > unknown
func (e *Extractor) ExtractFromHierarchy(resource *resourcepb.Resource, scope *commonpb.InstrumentationScope, logRecord *logspb.LogRecord) string {
	// First try log-level attributes (highest priority)
	if logRecord != nil {
		if value := e.ExtractValue(logRecord.Attributes); value != UnknownValue {
			return value
		}
	}

	// Then try scope-level attributes
	if scope != nil {
		if value := e.ExtractValue(scope.Attributes); value != UnknownValue {
			return value
		}
	}

	// Finally try resource-level attributes
	if resource != nil {
		if value := e.ExtractValue(resource.Attributes); value != UnknownValue {
			return value
		}
	}

	// Return unknown if not found at any level
	return UnknownValue
}

// ExtractValue extracts the attribute value from a list of KeyValue pairs
// Returns UnknownValue if the attribute is not found
func (e *Extractor) ExtractValue(attributes []*commonpb.KeyValue) string {
	for _, attr := range attributes {
		if attr.Key == e.attributeKey {
			return e.getStringValue(attr.Value)
		}
	}
	return UnknownValue
}

// arrayToSlice converts ArrayValue to Go slice
func (e *Extractor) arrayToSlice(arr *commonpb.ArrayValue) []interface{} {
	if arr == nil {
		return nil
	}
	result := make([]interface{}, 0, len(arr.Values))
	for _, v := range arr.Values {
		result = append(result, e.anyValueToInterface(v))
	}
	return result
}

// kvListToMap converts KeyValueList to Go map
func (e *Extractor) kvListToMap(kvList *commonpb.KeyValueList) map[string]interface{} {
	if kvList == nil {
		return nil
	}
	result := make(map[string]interface{})
	for _, kv := range kvList.Values {
		result[kv.Key] = e.anyValueToInterface(kv.Value)
	}
	return result
}

// anyValueToInterface converts AnyValue to Go interface{}
func (e *Extractor) anyValueToInterface(value *commonpb.AnyValue) interface{} {
	if value == nil {
		return nil
	}
	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_IntValue:
		return v.IntValue
	case *commonpb.AnyValue_DoubleValue:
		return v.DoubleValue
	case *commonpb.AnyValue_BoolValue:
		return v.BoolValue
	case *commonpb.AnyValue_ArrayValue:
		return e.arrayToSlice(v.ArrayValue)
	case *commonpb.AnyValue_KvlistValue:
		return e.kvListToMap(v.KvlistValue)
	case *commonpb.AnyValue_BytesValue:
		return fmt.Sprintf("%x", v.BytesValue)
	default:
		return nil
	}
}

// getStringValue converts an AnyValue to a string representation
func (e *Extractor) getStringValue(value *commonpb.AnyValue) string {
	if value == nil {
		return UnknownValue
	}

	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_IntValue:
		return fmt.Sprintf("%d", v.IntValue)
	case *commonpb.AnyValue_DoubleValue:
		return fmt.Sprintf("%f", v.DoubleValue)
	case *commonpb.AnyValue_BoolValue:
		return fmt.Sprintf("%t", v.BoolValue)
	case *commonpb.AnyValue_ArrayValue:
		// Serialize array to JSON string for better visibility
		if data, err := json.Marshal(e.arrayToSlice(v.ArrayValue)); err == nil {
			return string(data)
		}
		return UnknownValue
	case *commonpb.AnyValue_KvlistValue:
		// Serialize map to JSON string for better visibility
		if data, err := json.Marshal(e.kvListToMap(v.KvlistValue)); err == nil {
			return string(data)
		}
		return UnknownValue
	case *commonpb.AnyValue_BytesValue:
		// Encode bytes as base64 string
		return fmt.Sprintf("base64:%x", v.BytesValue)
	default:
		return UnknownValue
	}
}

package attributes

import (
	"testing"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

func TestExtractor_ExtractValue(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		attributes []*commonpb.KeyValue
		want       string
	}{
		{
			name: "string value found",
			key:  "service.name",
			attributes: []*commonpb.KeyValue{
				{
					Key: "service.name",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "my-service"},
					},
				},
			},
			want: "my-service",
		},
		{
			name: "int value found",
			key:  "status.code",
			attributes: []*commonpb.KeyValue{
				{
					Key: "status.code",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_IntValue{IntValue: 200},
					},
				},
			},
			want: "200",
		},
		{
			name: "bool value found",
			key:  "is_error",
			attributes: []*commonpb.KeyValue{
				{
					Key: "is_error",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_BoolValue{BoolValue: true},
					},
				},
			},
			want: "true",
		},
		{
			name: "double value found",
			key:  "duration",
			attributes: []*commonpb.KeyValue{
				{
					Key: "duration",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_DoubleValue{DoubleValue: 1.234},
					},
				},
			},
			want: "1.234000",
		},
		{
			name: "key not found",
			key:  "missing.key",
			attributes: []*commonpb.KeyValue{
				{
					Key: "other.key",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "value"},
					},
				},
			},
			want: UnknownValue,
		},
		{
			name:       "empty attributes",
			key:        "any.key",
			attributes: []*commonpb.KeyValue{},
			want:       UnknownValue,
		},
		{
			name:       "nil attributes",
			key:        "any.key",
			attributes: nil,
			want:       UnknownValue,
		},
		{
			name: "multiple attributes, find correct one",
			key:  "foo",
			attributes: []*commonpb.KeyValue{
				{
					Key: "bar",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "wrong"},
					},
				},
				{
					Key: "foo",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "correct"},
					},
				},
				{
					Key: "baz",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{StringValue: "also-wrong"},
					},
				},
			},
			want: "correct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExtractor(tt.key)
			got := e.ExtractValue(tt.attributes)
			if got != tt.want {
				t.Errorf("ExtractValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractor_getStringValue(t *testing.T) {
	e := NewExtractor("test")

	tests := []struct {
		name  string
		value *commonpb.AnyValue
		want  string
	}{
		{
			name:  "nil value",
			value: nil,
			want:  UnknownValue,
		},
		{
			name: "array value serialized to JSON",
			value: &commonpb.AnyValue{
				Value: &commonpb.AnyValue_ArrayValue{
					ArrayValue: &commonpb.ArrayValue{},
				},
			},
			want: "[]",
		},
		{
			name: "array with values",
			value: &commonpb.AnyValue{
				Value: &commonpb.AnyValue_ArrayValue{
					ArrayValue: &commonpb.ArrayValue{
						Values: []*commonpb.AnyValue{
							{Value: &commonpb.AnyValue_StringValue{StringValue: "test"}},
						},
					},
				},
			},
			want: `["test"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.getStringValue(tt.value)
			if got != tt.want {
				t.Errorf("getStringValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractor_ComplexTypes(t *testing.T) {
	e := NewExtractor("test")

	// Test map/kvlist
	kvListValue := &commonpb.AnyValue{
		Value: &commonpb.AnyValue_KvlistValue{
			KvlistValue: &commonpb.KeyValueList{
				Values: []*commonpb.KeyValue{
					{
						Key: "key1",
						Value: &commonpb.AnyValue{
							Value: &commonpb.AnyValue_StringValue{StringValue: "value1"},
						},
					},
				},
			},
		},
	}

	result := e.getStringValue(kvListValue)
	expected := `{"key1":"value1"}`
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

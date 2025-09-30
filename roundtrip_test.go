package zon

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestEncoderDecoderRoundTrip(t *testing.T) {
	v := map[string]any{
		"name":   "Bob",
		"age":    42,
		"active": true,
	}

	var buf bytes.Buffer

	if err := NewEncoder(&buf).Encode(v); err != nil {
		t.Fatalf("Encoder.Encode failed: %v", err)
	}

	var out map[string]any

	if err := NewDecoder(&buf).Decode(&out); err != nil {
		t.Fatalf("Decoder.Decode failed: %v", err)
	}

	if !mapsDeepEqual(v, out) {
		t.Errorf("Encoder/Decoder round-trip mismatch\nexpected: %#v\nactual:   %#v", v, out)
	}
}

func mapsDeepEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	toInt64 := func(v any) int64 {
		switch x := v.(type) {
		case int:
			return int64(x)
		case int8:
			return int64(x)
		case int16:
			return int64(x)
		case int32:
			return int64(x)
		case int64:
			return x
		case float32:
			return int64(x)
		case float64:
			return int64(x)
		}

		panic(fmt.Sprintf("toInt64: unsupported type %T", v))
	}

	toUint64 := func(v any) uint64 {
		switch x := v.(type) {
		case uint:
			return uint64(x)
		case uint8:
			return uint64(x)
		case uint16:
			return uint64(x)
		case uint32:
			return uint64(x)
		case uint64:
			return x
		case float32:
			return uint64(x)
		case float64:
			return uint64(x)
		}

		panic(fmt.Sprintf("toUint64: unsupported type %T", v))
	}

	toFloat64 := func(v any) float64 {
		switch x := v.(type) {
		case float32:
			return float64(x)
		case float64:
			return x
		case int:
			return float64(x)
		case int8:
			return float64(x)
		case int16:
			return float64(x)
		case int32:
			return float64(x)
		case int64:
			return float64(x)
		case uint:
			return float64(x)
		case uint8:
			return float64(x)
		case uint16:
			return float64(x)
		case uint32:
			return float64(x)
		case uint64:
			return float64(x)
		}

		panic(fmt.Sprintf("toFloat64: unsupported type %T", v))
	}

	for k, v := range a {
		w, ok := b[k]
		if !ok {
			return false
		}

		switch x := v.(type) {
		case int, int8, int16, int32, int64:
			if toInt64(x) != toInt64(w) {
				return false
			}
		case uint, uint8, uint16, uint32, uint64:
			if toUint64(x) != toUint64(w) {
				return false
			}
		case float32, float64:
			if toFloat64(x) != toFloat64(w) {
				return false
			}
		default:
			if !reflect.DeepEqual(v, w) {
				return false
			}
		}
	}

	return true
}

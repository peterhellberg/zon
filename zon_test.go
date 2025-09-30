package zon

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"bool true", true},
		{"int", 42},
		{"float", 3.14},
		{"string", "hello"},
		{"slice", []int{1, 2, 3}},
		{"map", map[string]int{".a": 1}},
		{"struct", struct {
			X int `zon:"x"`
		}{X: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Marshal(tt.value)
			if err != nil {
				t.Errorf("Marshal(%v) returned error: %v", tt.value, err)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		data string
		out  interface{}
	}{
		{"bool true", "true", new(bool)},
		{"int", "42", new(int)},
		{"float", "3.14", new(float64)},
		{"string", `"hello"`, new(string)},
		{"slice", "[1,2,3]", &[]int{}},
		{"map", "{.a = 1}", &map[string]int{}},
		{"struct", "{.x = 5}", &struct {
			X int `zon:"x"`
		}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.data), tt.out)
			if err != nil {
				t.Errorf("Unmarshal(%q) returned error: %v", tt.data, err)
			}
			if reflect.ValueOf(tt.out).Elem().IsZero() {
				t.Errorf("Unmarshal did not set value for %q", tt.data)
			}
		})
	}
}

func TestEncoder(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input map[string]int
		want  string
	}{
		{
			name:  "key with dot",
			input: map[string]int{".a": 1},
			want:  "{.a = 1}",
		},
		{
			name:  "key without dot",
			input: map[string]int{"b": 2},
			want:  "{.b = 2}",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			enc := NewEncoder(&buf)

			if err := enc.Encode(tt.input); err != nil {
				t.Errorf("Encoder.Encode returned error: %v", err)
			}

			if buf.Len() == 0 {
				t.Error("Encoder.Encode wrote no data")
			}

			for k := range tt.input {
				if !strings.Contains(buf.String(), k) {
					t.Errorf("Encoded output does not contain key %q: %s", k, buf.String())
				}
			}

			if got := buf.String(); got != tt.want {
				t.Fatalf("buf.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecoder(t *testing.T) {
	b := []byte("{.a = 1}")
	dec := NewDecoder(bytes.NewReader(b))

	var out map[string]int
	if err := dec.Decode(&out); err != nil {
		t.Errorf("Decoder.Decode returned error: %v", err)
	}

	if out["a"] != 1 {
		t.Errorf("Decoder.Decode did not set correct value, got %+v", out)
	}
}

func TestEncoderDecoderRoundTrip(t *testing.T) {
	v := map[string]any{
		"name":   "Bob",
		"age":    42,
		"active": true,
	}

	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		t.Fatalf("Encoder.Encode failed: %v", err)
	}

	dec := NewDecoder(&buf)
	var out map[string]any
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("Decoder.Decode failed: %v", err)
	}

	normalizedOut := normalizeDecoded(out)

	if !mapsEqualNormalized(v, normalizedOut.(map[string]any)) {
		t.Errorf("Encoder/Decoder round-trip mismatch %#v %#v", v, normalizedOut)
	}
}

func normalizeDecoded(v any) any {
	switch vv := v.(type) {
	case map[string]any:
		nm := make(map[string]any, len(vv))
		for k, val := range vv {
			key := k
			if len(k) > 0 && k[0] == '.' {
				key = k[1:]
			}
			nm[key] = normalizeDecoded(val)
		}
		return nm
	case []any:
		nl := make([]any, len(vv))
		for i, val := range vv {
			nl[i] = normalizeDecoded(val)
		}
		return nl
	case float64:
		if float64(int(vv)) == vv {
			return int(vv)
		}
		return vv
	default:
		return vv
	}
}

func mapsEqualNormalized(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if !valuesEqualNormalized(va, vb) {
			return false
		}
	}
	return true
}

func valuesEqualNormalized(a, b any) bool {
	switch va := a.(type) {
	case int, int8, int16, int32, int64:
		ai := reflect.ValueOf(va).Int()
		switch vb := b.(type) {
		case int, int8, int16, int32, int64:
			return ai == reflect.ValueOf(vb).Int()
		case float32, float64:
			return float64(ai) == reflect.ValueOf(vb).Float()
		}
	case uint, uint8, uint16, uint32, uint64:
		ui := reflect.ValueOf(va).Uint()
		switch vb := b.(type) {
		case uint, uint8, uint16, uint32, uint64:
			return ui == reflect.ValueOf(vb).Uint()
		case float32, float64:
			return float64(ui) == reflect.ValueOf(vb).Float()
		}
	case float32, float64:
		af := reflect.ValueOf(va).Float()
		switch vb := b.(type) {
		case float32, float64:
			return af == reflect.ValueOf(vb).Float()
		case int, int8, int16, int32, int64:
			return af == float64(reflect.ValueOf(vb).Int())
		case uint, uint8, uint16, uint32, uint64:
			return af == float64(reflect.ValueOf(vb).Uint())
		}
	case string, bool, nil:
		return reflect.DeepEqual(a, b)
	case map[string]any:
		vbMap, ok := b.(map[string]any)
		if !ok {
			return false
		}
		return mapsEqualNormalized(va, vbMap)
	case []any:
		vbSlice, ok := b.([]any)
		if !ok {
			return false
		}
		if len(va) != len(vbSlice) {
			return false
		}
		for i := range va {
			if !valuesEqualNormalized(va[i], vbSlice[i]) {
				return false
			}
		}
		return true
	}
	return false
}

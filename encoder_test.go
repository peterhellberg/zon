package zon

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncoder(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input map[string]int
		want  string
	}{
		{"key with dot", map[string]int{".a": 1}, ".{.a = 1}"},
		{"key without dot", map[string]int{"b": 2}, ".{.b = 2}"},
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

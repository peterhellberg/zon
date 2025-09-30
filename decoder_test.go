package zon

import (
	"bytes"
	"testing"
)

func TestDecoder(t *testing.T) {
	b := []byte(".{.a = 1}")

	dec := NewDecoder(bytes.NewReader(b))

	var out map[string]int

	if err := dec.Decode(&out); err != nil {
		t.Errorf("Decoder.Decode returned error: %v", err)
	}

	if out["a"] != 1 {
		t.Errorf("Decoder.Decode did not set correct value, got %+v", out)
	}
}

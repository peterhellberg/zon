package zon

import (
	"bytes"
	"testing"
)

func TestDecoder(t *testing.T) {
	b := []byte(".{.a = 1}")

	dec := NewDecoder(bytes.NewReader(b))

	var v map[string]int

	if err := dec.Decode(&v); err != nil {
		t.Errorf("Decoder.Decode returned error: %v", err)
	}

	if v["a"] != 1 {
		t.Errorf("Decoder.Decode did not set correct value, got %+v", v)
	}
}

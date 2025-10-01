package zon_test

import (
	"bytes"
	"testing"

	"github.com/peterhellberg/zon"
)

func FuzzStructRoundTrip(f *testing.F) {
	f.Add([]byte(`.{.name="Alice",.age=30,.active=true,.meta={.nested=.{1,2,3},.flag=true}}`))
	f.Add([]byte(`.{.name="Bob",.age=42,.active=false,.meta=null}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var v map[string]any

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Unmarshal panicked on input: %q, panic: %v", data, r)
			}
		}()

		_ = zon.Unmarshal(data, &v)
	})
}

func FuzzMapRoundTrip(f *testing.F) {
	f.Add([]byte(`.{.name="Peter",.age=42}`))
	f.Add([]byte(`.{.foo="bar",.baz=.{1,2,3}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var v map[string]any

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Unmarshal panicked on input: %q, panic: %v", data, r)
			}
		}()

		if err := zon.Unmarshal(data, &v); err != nil {
			t.Logf("Unmarshal returned error: %v", err)
			return
		}

		if _, err := zon.Marshal(v); err != nil {
			t.Fatalf("Marshal after Unmarshal failed: %v", err)
		}
	})
}

func FuzzSliceRoundTrip(f *testing.F) {
	f.Add([]byte(`.{1,"two",true}`))
	f.Add([]byte(`.{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var v []any

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Unmarshal panicked on input: %q, panic: %v", data, r)
			}
		}()

		if err := zon.Unmarshal(data, &v); err != nil {
			t.Logf("Unmarshal returned error: %v", err)
			return
		}

		if _, err := zon.Marshal(v); err != nil {
			t.Fatalf("Marshal after Unmarshal failed: %v", err)
		}
	})
}

func FuzzEncoderDecoder(f *testing.F) {
	f.Add([]byte(`.{.name="Fuzz",.age=99,.active=true}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var buf bytes.Buffer

		{
			var v map[string]any

			if err := zon.Unmarshal(data, &v); err != nil {
				return
			}

			if err := zon.Encode(&buf, v); err != nil {
				t.Fatalf("Encoder failed: %v", err)
			}
		}

		var v map[string]any

		if err := zon.Decode(&buf, &v); err != nil {
			t.Fatalf("Decoder failed: %v", err)
		}
	})
}

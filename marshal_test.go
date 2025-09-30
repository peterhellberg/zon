package zon

import "testing"

func TestMarshal(t *testing.T) {
	for _, tt := range []struct {
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
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Marshal(tt.value)
			if err != nil {
				t.Errorf("Marshal(%v) returned error: %v", tt.value, err)
			}
		})
	}
}

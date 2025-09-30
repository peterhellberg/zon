package zon

import (
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	for _, tt := range []struct {
		name string
		data string
		out  interface{}
	}{
		{"bool true", "true", new(bool)},
		{"int", "42", new(int)},
		{"float", "3.14", new(float64)},
		{"string", `"hello"`, new(string)},
		{"slice", ".{1,2,3}", &[]int{}},
		{"map", ".{.a = 1}", &map[string]int{}},
		{"struct", ".{.x = 5, .y = .{.z = .{1,2}}}", &struct {
			X int `zon:"x"`
			Y struct {
				Z []int `zon:"z"`
			} `zon:"y"`
		}{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal([]byte(tt.data), tt.out); err != nil {
				t.Errorf("Unmarshal(%q) returned error: %v", tt.data, err)
			}

			if reflect.ValueOf(tt.out).Elem().IsZero() {
				t.Errorf("Unmarshal did not set value for %q", tt.data)
			}
		})
	}
}

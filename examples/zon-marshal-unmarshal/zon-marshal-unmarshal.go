package main

import (
	"fmt"

	"github.com/peterhellberg/zon"
)

type Example struct {
	Name string `zon:"name"`
	Age  int    `zon:"age"`
	List []int  `zon:"list"`
	Omit []int  `zon:"omit,omitempty"`
}

func main() {
	v := Example{Name: "Peter", Age: 42}

	if err := run(v); err != nil {
		panic(err)
	}
}

func run(v Example) error {
	data, err := zon.Marshal(v, zon.Indent(""))
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	// Output: .{ .name = "Peter", .age = 42, .list = .{ }, }

	var v2 map[string]any

	if err := zon.Unmarshal(data, &v2); err != nil {
		return err
	}

	fmt.Printf("%+v\n", v2)
	// Output: map[age:42 list:[] name:Peter]

	return nil
}

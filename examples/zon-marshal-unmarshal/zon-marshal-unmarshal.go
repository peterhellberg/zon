package main

import (
	"fmt"

	"github.com/peterhellberg/zon"
)

type Example struct {
	Name string `zon:"name"`
	Age  int    `zon:"age"`
}

func main() {
	v := Example{Name: "Peter", Age: 42}

	if err := run(v); err != nil {
		panic(err)
	}
}

func run(v Example) error {
	data, err := zon.Marshal(v)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	// Output: .{.name = "Peter", .age = 42}

	var out map[string]any

	if err := zon.Unmarshal(data, &out); err != nil {
		return err
	}

	fmt.Printf("%+v\n", out)
	// Output: map[age:42 name:Peter]

	return nil
}

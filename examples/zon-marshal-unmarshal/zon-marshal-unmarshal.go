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

	// Marshal to Zon
	data, err := zon.Marshal(v)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	// Unmarshal from Zon into a map
	var out map[string]any

	if err := zon.Unmarshal(data, &out); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", out)
}

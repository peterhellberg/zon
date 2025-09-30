package main

import (
	"bytes"
	"fmt"

	"github.com/peterhellberg/zon"
)

type Example struct {
	Name string `zon:"name"`
}

func main() {
	v := Example{Name: "Peter"}

	if err := run(v); err != nil {
		panic(err)
	}
}

func run(v Example) error {
	var buf bytes.Buffer

	if err := zon.Encode(&buf, v); err != nil {
		return err
	}

	fmt.Println(buf.String())
	// Output: {.name = "Peter"}

	var out Example

	if err := zon.Decode(&buf, &out); err != nil {
		return err
	}

	fmt.Printf("%+v\n", out)
	// Output: {Name:Peter}

	return nil
}

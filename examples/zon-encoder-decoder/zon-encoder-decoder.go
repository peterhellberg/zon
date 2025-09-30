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

	var buf bytes.Buffer

	enc := zon.NewEncoder(&buf)

	if err := enc.Encode(v); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())

	var out Example

	dec := zon.NewDecoder(&buf)

	if err := dec.Decode(&out); err != nil {
		panic(err)
	}

	fmt.Println(out)
}

# zon - Zig Zon encoding/decoding for Go ⚡

`zon` is a Go library for marshaling and unmarshaling [Zig Zon](https://ziglang.org/) data,
similar in usage to `encoding/json`.

## Features

- Marshal Go primitives and structs into Zig Zon format
- Unmarshal Zig Zon data into Go values
- Support for `Encoder` and `Decoder`
- Handles booleans, numbers, strings, slices, maps, and structs

## Installation

```console
$ go get github.com/peterhellberg/zon
```

## Usage

### Marshal / Unmarshal

[embedmd]:# (examples/zon-marshal-unmarshal/zon-marshal-unmarshal.go)
```go
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
	v := Example{
		Name: "Alice",
		Age:  30,
	}

	// Marshal to Zon
	data, err := zon.Marshal(v)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	// Unmarshal from Zon
	var out Example

	if err := zon.Unmarshal(data, &out); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", out)
}
```

### Encoder / Decoder

[embedmd]:# (examples/zon-encoder-decoder/zon-encoder-decoder.go)
```go
package main

import (
	"bytes"

	"github.com/peterhellberg/zon"
)

type Example struct {
	Name string `zon:"name"`
}

func main() {
	var buf bytes.Buffer

	enc := zon.NewEncoder(&buf)
	dec := zon.NewDecoder(&buf)

	v := Example{Name: "Bob"}

	if err := enc.Encode(v); err != nil {
		panic(err)
	}

	var out Example

	if err := dec.Decode(&out); err != nil {
		panic(err)
	}
}
```

## License

MIT License

[embedmd]:# (LICENSE text)
```text
Copyright © 2025 Peter Hellberg - https://c7.se/

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

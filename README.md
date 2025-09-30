# ZON encoding/decoding for Go ⚡

`zon` is a Go library for marshaling and unmarshaling [ZON](https://ziglang.org/) data,
similar in usage to `encoding/json`.

> [!IMPORTANT]
> This library is not yet battle-tested; consider it more of an experiment :)

## Features

- Marshal Go primitives and structs into ZON format
- Unmarshal ZON data into Go values
- Support for `Encoder` and `Decoder`
- Handles booleans, numbers, strings, slices, maps, and structs

## Installation

    go get -u github.com/peterhellberg/zon

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
	data, err := zon.Marshal(v)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	// Output: .{.name = "Peter", .age = 42, .list = .{}}

	var out map[string]any

	if err := zon.Unmarshal(data, &out); err != nil {
		return err
	}

	fmt.Printf("%+v\n", out)
	// Output: map[age:42 list:[] name:Peter]

	return nil
}
```

### Encoder / Decoder

[embedmd]:# (examples/zon-encoder-decoder/zon-encoder-decoder.go)
```go
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

	if err := zon.NewEncoder(&buf).Encode(v); err != nil {
		return err
	}

	fmt.Println(buf.String())
	// Output: .{.name = "Peter"}

	var out Example

	if err := zon.NewDecoder(&buf).Decode(&out); err != nil {
		return err
	}

	fmt.Printf("%+v\n", out)
	// Output: {Name:Peter}

	return nil
}
```

> [!TIP]
> As a slight convenience, there are `zon.Encode` and `zon.Decode` functions;

[embedmd]:# (examples/zon-encoder-decoder-convenience/zon-encoder-decoder-convenience.go /func run/ $)
```go
func run(v Example) error {
	var buf bytes.Buffer

	if err := zon.Encode(&buf, v); err != nil {
		return err
	}

	fmt.Println(buf.String())
	// Output: .{.name = "Peter"}

	var out Example

	if err := zon.Decode(&buf, &out); err != nil {
		return err
	}

	fmt.Printf("%+v\n", out)
	// Output: {Name:Peter}

	return nil
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

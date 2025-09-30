package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/peterhellberg/zon"
)

type Encoder interface{ Encode(v any) error }
type Decoder interface{ Decode(v any) error }

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(r io.Reader, w io.Writer) error {
	j := flag.Bool("json", false, "")

	flag.Parse()

	if *j {
		return convert(zon.NewDecoder(r), json.NewEncoder(w))
	}

	return convert(json.NewDecoder(r), zon.NewEncoder(w))
}

func convert(dec Decoder, enc Encoder) error {
	var v any

	if err := dec.Decode(&v); err != nil {
		return err
	}

	return enc.Encode(v)
}

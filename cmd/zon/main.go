package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/peterhellberg/zon"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(r io.Reader, w io.Writer) error {
	i := flag.String("i", "    ", "Indentation per level")
	j := flag.Bool("j", false, "Convert ZON to JSON (default: false)")

	flag.Parse()

	if *j {
		dec := zon.NewDecoder(r)
		enc := json.NewEncoder(w)

		enc.SetIndent("", *i)

		return convert(dec, enc)
	}

	return convert(json.NewDecoder(r), zon.NewEncoder(w, zon.Indent(*i)))
}

type Decoder interface{ Decode(v any) error }
type Encoder interface{ Encode(v any) error }

func convert(dec Decoder, enc Encoder) error {
	var v any

	if err := dec.Decode(&v); err != nil {
		return err
	}

	return enc.Encode(v)
}

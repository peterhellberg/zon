package zon

import (
	"bytes"
	"io"
)

type Decoder struct{ r io.Reader }

func NewDecoder(r io.Reader) *Decoder { return &Decoder{r: r} }

func Decode(r io.Reader, out interface{}) error {
	return NewDecoder(r).Decode(out)
}

func (d *Decoder) Decode(out interface{}) error {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(d.r)
	if err != nil {
		return err
	}
	return Unmarshal(buf.Bytes(), out)
}

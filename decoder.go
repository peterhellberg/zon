package zon

import (
	"bytes"
	"io"
)

type Decoder struct {
	r io.Reader
}

func Decode(r io.Reader, v any) error {
	return NewDecoder(r).Decode(v)
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v any) error {
	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(d.r); err != nil {
		return err
	}

	return Unmarshal(buf.Bytes(), v)
}

package zon

import "io"

type Encoder struct {
	w io.Writer
	o []Option
}

func Encode(w io.Writer, v any, opts ...Option) error {
	return NewEncoder(w, opts...).Encode(v)
}

func NewEncoder(w io.Writer, opts ...Option) *Encoder {
	return &Encoder{w: w, o: opts}
}

func (e *Encoder) Encode(v any) error {
	data, err := Marshal(v, e.o...)
	if err != nil {
		return err
	}

	_, err = e.w.Write(data)

	return err
}

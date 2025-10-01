package zon

func defaultOptions() Options {
	return Options{
		Indent: "    ",
	}
}

type Options struct {
	Indent string
}

type Option func(o *Options)

func Indent(s string) Option {
	return func(o *Options) {
		o.Indent = s
	}
}

// Package zon provides serialization and deserialization of the ZON data format for Go values.
//
// The package allows converting Go values to ZON and back, supporting basic types
// (bool, int, uint, float, string), slices, arrays, maps, structs, pointers, and interfaces.
//
// Key features:
//
//   - Marshal, Unmarshal, Encode and Decode functions.
//   - Encoder and Decoder types.
//   - Support for struct field tags via `zon:"name"` to customize serialized field names.
//   - Map keys are automatically prefixed with a dot (`.`) unless already present.
//   - Pointers and interface values are handled transparently, with `nil` encoded as `null`.
//   - Graceful handling of unknown fields during struct deserialization.
//   - A simple syntax that uses `.`-prefixed field keys and `=` as a key-value separator.
package zon

//go:generate go tool github.com/campoy/embedmd -w README.md

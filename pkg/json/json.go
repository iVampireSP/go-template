package json

import (
	stdjson "encoding/json"
	"io"

	"github.com/bytedance/sonic"
)

// Re-export encoding/json types so callers need not import encoding/json directly.
type (
	RawMessage = stdjson.RawMessage
	Number     = stdjson.Number
)

// Marshal encodes v to JSON.
func Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

// MarshalIndent is like Marshal but applies indent formatting.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return sonic.MarshalIndent(v, prefix, indent)
}

// Unmarshal decodes JSON bytes into v.
func Unmarshal(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

// UnmarshalString decodes a JSON string into v.
func UnmarshalString(s string, v any) error {
	return sonic.UnmarshalString(s, v)
}

// Valid reports whether data is valid JSON.
func Valid(data []byte) bool {
	return sonic.Valid(data)
}

// Encoder wraps a sonic streaming encoder.
type Encoder struct {
	enc sonic.Encoder
}

// NewEncoder returns a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		sonic.ConfigDefault.NewEncoder(w),
	}
}

// Encode writes the JSON encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	return e.enc.Encode(v)
}

// SetIndent sets the indentation for the encoder output.
func (e *Encoder) SetIndent(prefix, indent string) {
	e.enc.SetIndent(prefix, indent)
}

// Decoder wraps a sonic streaming decoder.
type Decoder struct {
	dec sonic.Decoder
}

// NewDecoder returns a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		sonic.ConfigDefault.NewDecoder(r),
	}
}

// Decode reads the next JSON value from the stream and stores it in v.
func (d *Decoder) Decode(v any) error {
	return d.dec.Decode(v)
}

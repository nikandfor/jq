package jqgo

import (
	"encoding/json"

	"nikand.dev/go/jq"
	"nikand.dev/go/jq/jqjson"
)

type (
	Encoder struct {
		Reference any

		buf []byte
	}
)

func NewEncoder(ref any) *Encoder {
	return &Encoder{Reference: ref}
}

func (e *Encoder) ApplyTo(b *jq.Buffer, off jq.Off, next bool) (jq.Off, bool, error) {
	if next {
		return jq.None, false, nil
	}

	_, err := e.Encode(nil, b, off)
	if err != nil {
		return jq.None, false, err
	}

	return off, false, nil
}

// Encode ignores w and unmarshals data into e.Reference.
func (e *Encoder) Encode(w []byte, b *jq.Buffer, off jq.Off) (_ []byte, err error) {
	jse := jqjson.NewEncoder()

	e.buf, err = jse.Encode(e.buf[:0], b, off)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(e.buf, e.Reference)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

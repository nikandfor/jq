package jqgo

import (
	"bytes"
	"encoding/json"

	"nikand.dev/go/jq"
	"nikand.dev/go/jq/jqjson"
)

type (
	Decoder struct {
		Value any

		buf bytes.Buffer
	}
)

func NewDecoder(val any) *Decoder {
	return &Decoder{Value: val}
}

func (d *Decoder) ApplyTo(b *jq.Buffer, off jq.Off, next bool) (jq.Off, bool, error) {
	if next {
		return jq.None, false, nil
	}

	off, _, err := d.Decode(b, nil, 0)
	return off, false, err
}

func (d *Decoder) Decode(b *jq.Buffer, r []byte, st int) (off jq.Off, i int, err error) {
	d.buf.Reset()

	e := json.NewEncoder(&d.buf)

	err = e.Encode(d.Value)
	if err != nil {
		return jq.None, 0, err
	}

	jqd := jqjson.NewDecoder()

	return jqd.Decode(b, d.buf.Bytes(), 0)
}

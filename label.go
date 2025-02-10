package jq

import (
	"errors"

	"nikand.dev/go/cbor"
)

type (
	Label struct{}
)

var errInvaliLabel = errors.New("invalid label")

func NewLabel() Label { return Label{} }

func (f Label) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	if off < 0 {
		return Null, false, nil
	}

	br := b.Reader()
	bw := b.Writer()

	tag := br.Tag(off)
	if tag != cbor.Label {
		return Null, false, nil
	}

	lab := br.Label(off)
	if lab < LabelOffset {
		return off, false, errInvaliLabel
	}

	res = bw.Int(lab - LabelOffset)

	return res, false, nil
}

func (f Label) String() string {
	return "label"
}

package eval

import (
	"fmt"

	"github.com/nikandfor/errors"
)

type (
	JSON struct{}

	KVPair struct {
		K String
		V any

		Sep Parser
	}

	Null struct{}

	Bool struct{}
)

func (JSON) Parse(p []byte, st int) (any, int, error) {
	return Spacer{
		Spaces: GoSpaces,
		Of: AnyOf{
			Null{},
			Bool{},
			Num{},
			String(nil),
			ListParser{
				Open: '[',
				Sep:  ',',
				Of:   JSON{},
			},
			ListParser{
				Open: '{',
				Sep:  ',',
				Of: KVPair{
					Sep: Spacer{Spaces: GoSpaces, Of: Const(":")},
				},
			},
		},
	}.Parse(p, st)
}

func (kv KVPair) Parse(p []byte, st int) (x any, i int, err error) {
	k, i, err := String{}.Parse(p, st)
	if err != nil {
		return nil, i, errors.Wrap(err, "key")
	}

	_, i, err = kv.Sep.Parse(p, i)
	if err != nil {
		return nil, i, err
	}

	v, i, err := JSON{}.Parse(p, i)
	if err != nil {
		return nil, i, errors.Wrap(err, "value")
	}

	return KVPair{
		K: []byte(k.(String)),
		V: v,
	}, i, nil
}

func (Null) Parse(p []byte, st int) (_ any, i int, err error) {
	_, i, err = Const("null").Parse(p, st)
	if err != nil {
		return
	}

	return Null{}, i, nil
}

func (Bool) Parse(p []byte, st int) (_ any, i int, err error) {
	_, i, err = Const("false").Parse(p, st)
	if err == nil {
		return false, i, nil
	}

	_, i, err = Const("true").Parse(p, st)
	if err == nil {
		return true, i, nil
	}

	return
}

func (JSON) String() string {
	return "JSON"
}

func (x KVPair) String() string {
	if x.K == nil && x.V == nil {
		return "kv_pair"
	}

	return fmt.Sprintf("%q: %v", x.K, x.V)
}

func (Null) String() string {
	return "null"
}

func (Bool) String() string {
	return "bool"
}

package jq

import (
	"fmt"
)

type (
	Variable struct {
		Name  string
		Value Off
	}

	Binding struct {
		Name  string
		Value Filter
	}

	Var string

	VarNotDefinedError string
)

func (b *Buffer) Bind(name string, off Off) {
	b.Vars = append(b.Vars, Variable{Name: name, Value: off})
}

func (b *Buffer) Var(name string) Off {
	for i := len(b.Vars) - 1; i >= 0; i-- {
		v := b.Vars[i]

		if v.Name == name {
			return v.Value
		}
	}

	return None
}

func Bind(name string, v Filter) *Binding {
	return &Binding{Name: name, Value: v}
}

func (f *Binding) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	res, more, err = f.Value.ApplyTo(b, off, next)
	if err != nil || res == None {
		return None, false, err
	}

	b.Bind(f.Name, res)

	return off, more, nil
}

func (f Var) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	//	log.Printf("b vars: %v", b.Vars)
	res = b.Var(string(f))
	if res == None {
		return None, false, VarNotDefinedError(f)
	}

	return res, false, nil
}

func (f *Binding) String() string { return fmt.Sprintf("%+v as $%s", f.Value, f.Name) }
func (f Var) String() string      { return "$" + string(f) }

func (e VarNotDefinedError) Error() string { return "variable is not defined: " + string(e) }

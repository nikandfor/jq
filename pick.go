package jq

import (
	"fmt"
)

type (
	Pick struct {
		Expr FilterPath

		merge
	}
)

func NewPick(e FilterPath) *Pick { return &Pick{Expr: e} }

func (f *Pick) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

	f.reset()

	for {
		res, f.path, next, err = f.Expr.ApplyToGetPath(b, off, f.path[:0], next)
		if err != nil {
			return off, false, err
		}
		if res == None && next {
			continue
		}
		if res == None {
			break
		}

		//	log.Printf("pick %v#%v", f.path, res)

		err = f.set(b, res, f.path)
		if err != nil {
			return None, false, err
		}

		//	log.Printf("tree\n%v", f.root)

		if !next {
			break
		}
	}

	res = f.render(b, f.root)

	return res, false, nil
}

func (p Pick) String() string { return fmt.Sprintf("pick(%v)", p.Expr) }

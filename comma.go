package jq

type (
	Comma struct {
		Filters []Filter

		j    int
		next bool
	}
)

func NewComma(fs ...Filter) *Comma {
	return &Comma{Filters: fs}
}

func (f *Comma) ApplyTo(b *Buffer, off int, next bool) (res int, more bool, err error) {
	if !next {
		f.j = 0
		f.next = false
	}

	res = None

	for f.j < len(f.Filters) {
		res, f.next, err = f.Filters[f.j].ApplyTo(b, off, f.next)
		if err != nil {
			return off, false, err
		}

		if !f.next {
			f.j++
		}

		if res != None {
			break
		}
	}

	return res, f.j < len(f.Filters), nil
}

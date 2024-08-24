package jq

type (
	Array struct {
		Of Filter

		arr []int
	}
)

func NewArray(of Filter) *Array { return &Array{Of: of} }

func (f *Array) ApplyTo(b *Buffer, off int, next bool) (res int, err error) {
	if next {
		return None, nil
	}

	bw := b.Writer()

	reset := bw.Offset()
	defer bw.ResetIfErr(reset, &err)

	f.arr = f.arr[:0]
	next = false

	for {
		sub, err := f.Of.ApplyTo(b, off, next)
		if err != nil {
			return off, err
		}
		if sub == None {
			break
		}

		f.arr = append(f.arr, sub)
		next = true
	}

	off = bw.Array(f.arr)

	return off, nil
}

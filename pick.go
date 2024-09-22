package jq

import (
	"errors"
	"fmt"
	"strings"

	"nikand.dev/go/cbor"
)

type (
	Pick struct {
		Expr FilterPath

		path NodePath
		arr  []Off

		root *node
		buf  node
	}

	node struct {
		off   Off
		child []*node
		add   []*node

		tag byte
		key Off
	}
)

func NewPick(e FilterPath) *Pick { return &Pick{Expr: e} }

func (f *Pick) ApplyTo(b *Buffer, off Off, next bool) (res Off, more bool, err error) {
	if next {
		return None, false, nil
	}

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

func (f *Pick) set(b *Buffer, res Off, p NodePath) error {
	if f.root == nil {
		f.buf = node{off: Null}
		f.root = &f.buf
	}

	br := b.Reader()
	n := f.root

	for _, p := range p {
		var tag byte

		if p.Off != Null {
			tag = br.Tag(p.Off)
		} else {
			tag = csel[byte](p.Key != None, cbor.Map, cbor.Array)
		}

		n.off = p.Off
		n.tag = tag

		if tag == cbor.Array && p.Index < 0 {
			return errors.New("pick: append negative array index")
		}
		if tag == cbor.Map && p.Index < 0 {
			n = f.addkey(b, n, p.Key)
			continue
		}

		n.child = ensure(n.child, p.Index)

		if n.child[p.Index] == nil {
			n.child[p.Index] = &node{off: Null, key: None}
		}

		n = n.child[p.Index]
	}

	n.off = res

	return nil
}

func (f *Pick) addkey(b *Buffer, n *node, key Off) *node {
	var c *node

	for _, cc := range n.add {
		if b.Equal(cc.key, key) {
			c = cc
			break
		}
	}

	if c == nil {
		c = &node{off: Null, tag: cbor.Map, key: key}
		n.add = append(n.add, c)
	}

	return c
}

func (f *Pick) render(b *Buffer, n *node) Off {
	if n == nil {
		return Null
	}

	if len(n.child)+len(n.add) == 0 {
		return n.off
	}

	br := b.Reader()

	reset := len(f.arr)
	defer func() { f.arr = f.arr[:reset] }()

	f.arr = resize(f.arr, reset+(len(n.child)+len(n.add))*csel(n.tag == cbor.Map, 2, 1))
	arr := f.arr[reset:]

	j := 0

	for i, c := range n.child {
		if n.tag == cbor.Map && c == nil {
			continue
		}

		if n.tag == cbor.Map {
			arr[j], _ = br.ArrayMapIndex(n.off, i)
			j++
		}

		arr[j] = f.render(b, c)
		j++
	}

	for _, c := range n.add {
		arr[j] = c.key
		j++
		arr[j] = f.render(b, c)
		j++
	}

	return b.Writer().ArrayMap(n.tag, arr[:j])
}

func (p Pick) String() string { return fmt.Sprintf("pick(%v)", p.Expr) }

func (n *node) String() string {
	var b strings.Builder

	n.dump(&b, 0, 0)

	return b.String()
}

func (n *node) dump(b *strings.Builder, i, d int) {
	fmt.Fprintf(b, "%s%4x: ", "                    "[:2*d], i)

	if n == nil {
		b.WriteString("nil\n")
		return
	}

	fmt.Fprintf(b, "%x  %v  (key %v)\n", n.tag, n.off, n.key)

	for i, c := range n.child {
		c.dump(b, i, d+1)
	}

	for _, c := range n.add {
		c.dump(b, -1, d+1)
	}
}

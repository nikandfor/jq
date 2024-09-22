package jq

import (
	"errors"
	"fmt"
	"strings"

	"nikand.dev/go/cbor"
)

type (
	merge struct {
		root *node

		path NodePath
		arr  []Off
	}

	node struct {
		off   Off
		child []*node

		tag byte
		key Off
	}
)

func (m *merge) reset() {
	if m.root == nil {
		m.root = m.new()
	}

	m.resetnode(m.root)
}

func (m *merge) resetnode(n *node) {
	if n == nil {
		return
	}

	for _, c := range n.child {
		m.resetnode(c)
	}

	n.off = None
	n.child = n.child[:0]
	n.tag = 0
	n.key = None
}

func (m *merge) render(b *Buffer, n *node) Off {
	if n == nil {
		return Null
	}
	if len(n.child) == 0 {
		return n.off
	}

	reset := len(m.arr)
	defer func() { m.arr = m.arr[:reset] }()

	m.arr = ensure(m.arr, reset+len(n.child)*csel(n.tag == cbor.Map, 2, 1))
	arr := m.arr[reset:]

	j := 0

	for _, c := range n.child {
		if n.tag == cbor.Map {
			arr[j] = c.key
			j++
		}

		arr[j] = m.render(b, c)
		j++
	}

	return b.Writer().ArrayMap(n.tag, arr[:j])
}

func (m *merge) set(b *Buffer, res Off, p NodePath) error {
	n := m.root

	for _, p := range p {
		n.off = p.Off
		n.tag = m.tag(b, p)

		if n.tag == cbor.Map {
			n = m.key(b, n, p.Key)
			continue
		}

		if p.Index < 0 {
			return errors.New("set negative array index")
		}

		n = m.index(n, p.Index)
	}

	n.off = res

	return nil
}

func (m *merge) key(b *Buffer, n *node, key Off) *node {
	for _, c := range n.child {
		if b.Equal(c.key, key) {
			return c
		}
	}

	c := m.new()
	c.key = key

	n.child = append(n.child, c)

	return c
}

func (m *merge) index(n *node, i int) *node {
	n.child = ensure(n.child, i)

	if n.child[i] == nil {
		n.child[i] = m.new()
	}

	return n.child[i]
}

func (m *merge) tag(b *Buffer, p NodePathSeg) byte {
	if p.Off >= 0 {
		return b.Reader().Tag(p.Off)
	}
	if p.Off != Null {
		panic(p.Off)
	}

	if p.Key != None {
		return cbor.Map
	}
	if p.Index >= 0 {
		return cbor.Array
	}

	panic(p)
}

func (m *merge) new() *node {
	return &node{off: Null, key: None}
}

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
}

package scanner

import (
	"errors"
	"fmt"
	"io"
	"log"
	"unicode/utf8"
)

type (
	Token rune

	Scanner struct {
		i, end int
		st     int

		b []byte

		r io.Reader

		Position

		Mode       uint64
		Whitespace uint64

		IsIdent func(r rune, prev []byte) bool
		IsInt   func(r rune, prev []byte) bool
		IsFloat func(r rune, prev []byte) bool
		IsOp    func(r rune, prev []byte) bool

		IsString  func(r rune, prev []byte) bool
		IsChar    func(r rune, prev []byte) bool
		IsComment func(r rune, prev []byte) bool

		err error
	}

	Position struct {
		Filename string
		Offset   int
		Line     int
		Column   int
	}
)

const (
	_ = -flagBase - iota
	Err
	Op
	Ident
	Int
	Float
	Char
	String
	Comment

	skipped

	flagBase = 256

	Skip = 16 // token + Skip for skipping token
)

func (s *Scanner) Reset(r io.Reader) {
	s.r = r
	s.Position = Position{}
	s.i, s.end = 0, 0
	s.err = nil

	s.Whitespace = 1<<' ' | 1<<'\t' | 1<<'\n' | 1<<'\r'
	s.Mode = SetModeSkip(Comment)

	s.IsIdent = isIdent
	s.IsInt = isInt
	s.IsFloat = isFloat
	s.IsOp = isOp

	s.IsComment = isComment
}

func (s *Scanner) TokenText() []byte {
	if s.err != nil {
		return nil
	}

	return s.b[s.st:s.i]
}

func (s *Scanner) Scan() (r Token) {
	st, i, end := s.st, s.i, s.end
	defer func() {
		log.Printf("Scan %3x %3x %3x  <=  %3x %3x %3x  -> %-8v %-12q %v", s.st, s.i, s.end, st, i, end, r, s.TokenText(), s.err)
	}()

again:
	r = s.Next()
	if r == '\ufeff' && s.Offset == 0 {
		r = s.Next()
	}

	if r < 0 {
		return
	}

	for s.Whitespace&(1<<r) != 0 {
		r = s.Next()
	}

	if r < 0 {
		return
	}

	s.st = s.i - utf8.RuneLen(rune(r))

	switch {
	case s.IsIdent(rune(r), nil):
		r = s.scan(r, Ident, s.IsIdent)
	case s.IsInt(rune(r), nil):
		r = s.scan(r, Int, s.IsInt)
	case s.IsFloat(rune(r), nil):
		r = s.scan(r, Float, s.IsFloat)
	case s.IsOp(rune(r), nil):
		r = s.scan(r, Op, s.IsOp)
	case s.IsString != nil && s.IsString(rune(r), nil):
		r = s.scan(r, String, s.IsString)
	case s.IsChar != nil && s.IsChar(rune(r), nil):
		r = s.scan(r, Char, s.IsChar)
	case s.IsComment != nil && s.IsComment(rune(r), nil):
		r = s.scan(r, Comment, s.IsComment)
	default:
		switch r {
		case '"', '`':
			r = s.scanString(r)
		case '\'':
			r = s.scanChar(r)
		case '/', '#':
			r = s.scanComment(r)
		default:
			return r
		}
	}

	if r == skipped {
		log.Printf("Scan %3x %3x %3x  <=  %3x %3x %3x  -> %v %q %v  <- skip i  <- skip itt", s.st, s.i, s.end, st, i, end, r, s.TokenText(), s.err)
		goto again
	}

	return r
}

func (s *Scanner) scan(r, tok Token, ok func(r rune, prev []byte) bool) Token {
	if Mode(s.Mode, tok, 0) {
		return r
	}

	for ok(rune(r), s.b[s.st:s.i]) {
		r = s.Next()
	}

	if r < 0 {
		return r
	}

	if Mode(s.Mode, tok, Skip) {
		return skipped
	}

	s.Unnext(r)

	return tok
}

func (s *Scanner) Peek() (r Token) {
	r = s.Next()
	if r == '\ufeff' && s.Offset == 0 {
		r = s.Next()
	}

	s.Unnext(r)

	return r
}

func (s *Scanner) Next() (r Token) {
	//	defer func(st, i, end int) {
	//		log.Printf("Next %3x %3x %3x  <=  %3x %3x %3x  -> %v %v", s.st, s.i, s.end, st, i, end, r, s.err)
	//	}(s.st, s.i, s.end)

	if s.i == s.end {
		r = s.read()
		if r < 0 {
			return
		}
	}

	r = Token(s.b[s.i])

	if r >= utf8.RuneSelf {
		panic("support utf8")
	}

	s.i++

	return
}

func (s *Scanner) read() (r Token) {
	var n int

	const keep = 4

	if s.st > keep {
		s.end -= copy(s.b[keep:], s.b[s.st:s.end])
		s.i -= s.st - keep
		s.st = keep
	}

	if s.end == len(s.b) {
		if len(s.b) == 0 {
			s.b = make([]byte, 2)
		} else {
			s.b = append(s.b, 0, 0, 0, 0, 0, 0, 0, 0)
			s.b = s.b[:cap(s.b)]
		}
	}

	n, s.err = s.r.Read(s.b[s.end:])
	s.end += n

	if s.err != nil {
		return Err
	}

	return 0
}

func (s *Scanner) Unnext(r Token) {
	if r < 0 {
		return
	}

	s.i -= utf8.RuneLen(rune(r))
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) scanString(r Token) Token {
	f := r

loop:
	for {
		r = s.Next()
		if r < 0 {
			return r
		}

		switch {
		case r == f:
			break loop
		case r == '\\' && f == '"':
			r = s.Next()
			if r < 0 {
				return r
			}

			continue
		case r == '\n' && f == '"':
			s.err = errors.New("end of line in string")
			return Err
		}
	}

	if Mode(s.Mode, String, Skip) {
		return skipped
	}

	return String
}

func (s *Scanner) scanChar(r Token) Token {
	f := r

	r = s.Next()
	if r < 0 {
		return r
	}

	switch r {
	case '\\':
		panic("implement it")
	default:
	}

	r = s.Next()
	if r < 0 {
		return r
	}

	if r != f {
		s.err = errors.New("bad char")
		return Err
	}

	if Mode(s.Mode, Char, Skip) {
		return skipped
	}

	return Char
}

func (s *Scanner) scanComment(r Token) Token {
	f := r

loop:
	for {
		r = s.Next()

		if r < 0 {
			return r
		}

		switch {
		case r == '\n' && f == '#':
			break loop
		case r == '\n' && f == '/' && s.i-s.st > 1 && s.b[s.st+1] == '/':
			break loop
		}

		if s.st-s.i >= 4 && s.b[s.st] == '/' && s.b[s.st+1] == '*' && s.b[s.i-2] == '*' && s.b[s.i-1] == '/' {
			break loop
		}
	}

	if Mode(s.Mode, Comment, Skip) {
		return skipped
	}

	return Comment
}

func isInt(r rune, prev []byte) bool {
	return r >= '0' && r <= '9' || len(prev) != 0 && (r == 'x' || r == 'o' || r == 'b')
}

func isFloat(r rune, prev []byte) bool {
	return r >= '0' && r <= '9' || r == '.' || len(prev) != 0 && (r == 'e' || r == 'E')
}

func isIdent(r rune, prev []byte) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' && len(prev) != 0 || r == '_'
}

func isOp(r rune, prev []byte) bool {
	switch r {
	case '.', '+', '-', '*', '/', '%', '^', '=', '|':
		return true
	}

	return false
}

func isComment(r rune, prev []byte) bool {
	switch len(prev) {
	case 0:
		return r == '/'
	case 1:
		return r == '/' || r == '*'
	default:
		if prev[1] == '/' {
			return r != '\n'
		}

		return len(prev) < 4 || prev[len(prev)-1] != '/' || prev[len(prev)-2] != '*'
	}
}

func Mode(m uint64, f, skip Token) bool {
	return m&(1<<uint64(-f-flagBase+skip)) != 0
}

func SetMode(f ...int) (r Token) {
	for _, f := range f {
		r |= 1 << uint(-f-flagBase)
	}
	return
}

func SetModeSkip(f ...int) (r uint64) {
	for _, f := range f {
		r |= 1 << uint(-f-flagBase+Skip)
	}
	return
}

func (r Token) String() string {
	if r < 0 {
		switch r {
		case 0:
			return "0"
		case -flagBase:
			return "base"
		case Err:
			return "Err"
		case Op:
			return "Op"
		case Ident:
			return "Ident"
		case Int:
			return "Int"
		case Float:
			return "Float"
		case Char:
			return "Char"
		case String:
			return "String"
		case Comment:
			return "Comment"
		case Skip:
			return "Skip"
		case skipped:
			return "skipped"
		}

		return fmt.Sprintf("-%x", -rune(r)-flagBase)
	}

	if r >= utf8.RuneSelf {
		return fmt.Sprintf("%x", rune(r))
	}

	return fmt.Sprintf("%q", rune(r))
}

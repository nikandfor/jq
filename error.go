package jq

import "fmt"

type (
	Error struct {
		Filter Filter
		Sub    int
		Off    Off
		Err    error
	}
)

func NewError(f Filter, off Off, err error) *Error {
	return &Error{Filter: f, Sub: -1, Off: off, Err: err}
}

func fe(f Filter, off Off, err error) error {
	if err == nil {
		return nil
	}

	return &Error{Filter: f, Sub: -1, Off: off, Err: err}
}

func fse(f Filter, sub int, off Off, err error) error {
	if err == nil {
		return nil
	}

	return &Error{Filter: f, Sub: sub, Off: off, Err: err}
}

func (e *Error) Error() string {
	sub := ""
	if e.Sub >= 0 {
		sub = fmt.Sprintf("[%d]", e.Sub)
	}

	return fmt.Sprintf("%T%s (%v): %v", e.Filter, sub, e.Off, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

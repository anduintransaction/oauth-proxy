package errors

import (
	"fmt"
	"runtime"
	"strings"
)

type Error struct {
	err   error
	stack []string
}

func NewError(err error, stack []string) *Error {
	if err == nil {
		return nil
	}
	return &Error{err, stack}
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) Underlying() error {
	return e.err
}

func (e *Error) GetStack() []string {
	return e.stack
}

func (e *Error) Stack() string {
	return strings.Join(e.stack, "\n")
}

func Errorf(format string, data ...interface{}) error {
	return wrap(fmt.Errorf(format, data...), 3)
}

func Wrap(err error) error {
	return wrap(err, 3)
}

func Is(e1 error, e2 error) bool {
	if e1 == e2 {
		return true
	}
	if e1, ok := e1.(*Error); ok {
		return Is(e1.err, e2)
	}
	if e2, ok := e2.(*Error); ok {
		return Is(e1, e2.err)
	}
	return false
}

func wrap(err error, skip int) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*Error); ok {
		return err
	}
	return &Error{err, StackTrace(skip)}
}

func StackTrace(skip int) []string {
	stack := []string{}
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d (0x%x)", file, line, pc))
	}
	return stack
}

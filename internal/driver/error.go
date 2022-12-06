package driver

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func NewError(err error, code string, message string) *Error {
	if err != nil {
		err = errors.Errorf("[%s]: %s", code, message)
	}

	return &Error{
		err:     err,
		Code:    code,
		Message: message,
	}
}

func NewErrorf(err error, code string, format string, v ...interface{}) *Error {
	return NewError(err, code, fmt.Sprintf(format, v...))
}

type Error struct {
	err     error
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s]: %s", e.Code, e.Message)
}

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "%s\n%+v\n", e.Error(), e.err)
		}
	case 's', 'g':
		_, _ = io.WriteString(s, e.Error())
	}
}

package driver

import (
	"fmt"
	"io"
)

func NewError(err error, code string, message string) *Error {
	return &Error{
		err:     err,
		Code:    code,
		Message: message,
	}
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

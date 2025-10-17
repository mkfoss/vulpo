package vulpo

import "fmt"

type Error struct {
	message string
	wrapped error
}

func NewError(message string) Error {
	return Error{message: message}
}

func NewErrorf(message string, args ...any) Error {
	return Error{message: fmt.Sprintf(message, args...)}
}

func (e Error) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s, wrapped error: %s", e.message, e.wrapped.Error())
	}
	return e.message
}

func (e Error) Unwrap() error {
	return e.wrapped
}

func (e Error) SetWrapped(err error) Error {
	e.wrapped = err
	return e
}

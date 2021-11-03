package controller

import "fmt"

type Error struct {
	Message string
	Child   error
}

func (e *Error) Error() string {
	return fmt.Sprint(e.Message + " <- " + e.Child.Error())
}

func NewError(message string, err error) error {
	return &Error{
		Message: message,
		Child:   err,
	}
}
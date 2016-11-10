package bind

import "fmt"

type Error struct {
	Message string
	Args    []interface{}
}

func (e *Error) GetMessage() string {
	return fmt.Sprintf(e.Message, e.Args...)
}

func (e *Error) Error() string {
	return e.GetMessage()
}

func (e *Error) String() string {
	return e.GetMessage()
}

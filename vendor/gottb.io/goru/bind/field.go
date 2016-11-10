package bind

type Field struct {
	value string
	error *Error
}

func (f *Field) Value() string {
	return f.value
}

func (f *Field) HasError() bool {
	return f.error != nil
}

func (f *Field) Error() *Error {
	return f.error
}

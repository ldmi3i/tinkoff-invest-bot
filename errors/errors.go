package errors

type msgError struct {
	m string
}

func (msg msgError) Error() string {
	return msg.m
}

func NewUnexpectedError(msg string) error {
	return msgError{msg}
}

type doubleSubErr struct {
	m string
}

func (msg doubleSubErr) Error() string {
	return msg.m
}

func NewDoubleSubErr(msg string) error {
	return doubleSubErr{msg}
}

// NewNotImplemented auxiliary - to mark not implemented
func NewNotImplemented() error {
	return NewUnexpectedError("Not implemented yet!")
}

func ConvertToError(r interface{}) error {
	switch tp := r.(type) {
	case string:
		return NewUnexpectedError(tp)
	case error:
		return tp
	default:
		return NewUnexpectedError("Unexpected error")
	}
}

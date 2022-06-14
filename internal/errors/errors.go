package errors

import "errors"

func NewUnexpectedError(msg string) error {
	return errors.New(msg)
}

// NewNotImplemented auxiliary - to mark not implemented
func NewNotImplemented() error {
	return NewUnexpectedError("Not implemented yet!")
}

// ConvertToError converts recovered type to error
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

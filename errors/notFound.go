package errors

type NotFoundError struct {
	msg string
}

func (err NotFoundError) Error() string {
	return err.msg
}

func NewNotFound(msg string) NotFoundError {
	return NotFoundError{msg: msg}
}

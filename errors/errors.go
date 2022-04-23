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

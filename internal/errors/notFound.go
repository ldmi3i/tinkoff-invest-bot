package errors

//NotFoundErr happens when some resources not found
type NotFoundErr struct {
	msg string
}

func (err NotFoundErr) Error() string {
	return err.msg
}

func NewNotFound(msg string) NotFoundErr {
	return NotFoundErr{msg: msg}
}

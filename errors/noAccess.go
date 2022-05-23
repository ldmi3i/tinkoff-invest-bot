package errors

//NoAccessErr happens when not enough access rights to account provided by token
type NoAccessErr struct {
	msg string
}

func (err NoAccessErr) Error() string {
	return err.msg
}

func NewNoAccess(msg string) NoAccessErr {
	return NoAccessErr{msg: msg}
}

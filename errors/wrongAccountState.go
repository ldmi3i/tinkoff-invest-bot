package errors

//WrongAccStateErr happens when account in wrong state
type WrongAccStateErr struct {
	msg string
}

func (err WrongAccStateErr) Error() string {
	return err.msg
}

func NewWrongAccState(msg string) WrongAccStateErr {
	return WrongAccStateErr{msg: msg}
}

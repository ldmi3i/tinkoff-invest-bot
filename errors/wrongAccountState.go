package errors

type WrongAccState struct {
	msg string
}

func (err WrongAccState) Error() string {
	return err.msg
}

func NewWrongAccState(msg string) WrongAccState {
	return WrongAccState{msg: msg}
}

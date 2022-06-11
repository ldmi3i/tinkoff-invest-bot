package errors

//DoubleSubErr happens when multiple subscriptions not available, but called
type DoubleSubErr struct {
	m string
}

func (msg DoubleSubErr) Error() string {
	return msg.m
}

func NewDoubleSubErr(msg string) error {
	return DoubleSubErr{msg}
}

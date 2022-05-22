package errors

type NoAccess struct {
	msg string
}

func (err NoAccess) Error() string {
	return err.msg
}

func NewNoAccess(msg string) NoAccess {
	return NoAccess{msg: msg}
}

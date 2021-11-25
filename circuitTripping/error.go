package circuitTripping

type Error struct {
	Err error
}

func New(e error) *Error {
	return &Error{
		Err: e,
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func IsTripping(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*Error)
	return ok
}

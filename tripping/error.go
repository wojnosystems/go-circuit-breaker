package tripping

type Error struct {
	Err  error
	Cost uint64
}

// New converts your error into a tripping error, one the circuit breaker
// will recognize as an event which could trip the breaker
// err is assumed to have a unit cost of 1
func New(err error) *Error {
	return NewWithCost(err, 1)
}

// NewWithCost converts your error to a tripping error and assigns a tripping cost to the failure.
// Use this to assign a weight to this error
func NewWithCost(err error, cost uint64) *Error {
	return &Error{
		Err:  err,
		Cost: cost,
	}
}

// Error satisfies the Error interface by returning the wrapped error's string
func (e *Error) Error() string {
	return e.Err.Error()
}

// IsTripping evaluates the error or nil and returns true if this is a tripping error, or false if nil or some other error type
func IsTripping(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*Error)
	return ok
}

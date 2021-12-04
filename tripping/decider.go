package tripping

// Decider returns true if the breaker should trip and transition to the open state, false to stay in the closed state
// trippingErr is the error that caused the breaker to test the TripDecider
type Decider func(trippingErr *Error) (shouldTrip bool)

// ShouldTrip returns true if the trippingError should cause the breaker to move from a closed to an open state
func (t Decider) ShouldTrip(trippingErr *Error) bool {
	if t != nil {
		return t(trippingErr)
	}
	return true
}

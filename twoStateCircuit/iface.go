package twoStateCircuit

import (
	"time"
)

type nowFactory func() time.Time

type breaker interface {
	Use(callback func() error) error
}

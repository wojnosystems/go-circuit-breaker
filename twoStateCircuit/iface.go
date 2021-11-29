package twoStateCircuit

type breaker interface {
	Use(callback func() error) error
}

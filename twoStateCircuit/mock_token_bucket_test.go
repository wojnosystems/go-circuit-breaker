package twoStateCircuit

import "github.com/wojnosystems/go-circuit-breaker/tripping"

func neverTrips(_ *tripping.Error) bool {
	return false
}

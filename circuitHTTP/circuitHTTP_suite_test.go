package circuitHTTP_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCircuitHTTP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CircuitHTTP Suite")
}

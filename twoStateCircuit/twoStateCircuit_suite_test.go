package twoStateCircuit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTwoStateCircuit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TwoStateCircuit Suite")
}

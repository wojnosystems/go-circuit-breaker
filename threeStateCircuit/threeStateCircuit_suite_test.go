package threeStateCircuit

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestThreeStateCircuit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ThreeStateCircuit Suite")
}

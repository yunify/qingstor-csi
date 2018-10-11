package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSuit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(GinkgoT(), "Util Suite")
}

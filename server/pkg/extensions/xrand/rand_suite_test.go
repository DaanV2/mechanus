package xrand_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Random Suite")
}

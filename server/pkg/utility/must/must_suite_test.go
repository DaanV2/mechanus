package must_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMust(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Must Suite")
}

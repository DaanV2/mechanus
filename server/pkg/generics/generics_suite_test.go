package generics_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGenerics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generics Suite")
}

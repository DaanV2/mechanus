package authentication_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Testauthentication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "authentication Suite")
}

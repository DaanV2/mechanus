package screens_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScreens(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Screens Suite")
}

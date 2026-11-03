package screen_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScreen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Screen driver Suite")
}

package xsync_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestXsync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Xsync Suite")
}

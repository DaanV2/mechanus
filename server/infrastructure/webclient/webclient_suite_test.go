package webclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWebclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webclient Suite")
}

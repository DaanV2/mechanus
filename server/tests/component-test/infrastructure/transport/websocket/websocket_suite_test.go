package websocket_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test_Websocket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "websocket Suite")
}

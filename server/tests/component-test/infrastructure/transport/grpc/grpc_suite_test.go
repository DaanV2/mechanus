package grpc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Test_GRPC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "grpc Suite")
}

package authenication_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAuthenication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authenication Suite")
}

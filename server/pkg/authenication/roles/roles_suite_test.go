package roles_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRoles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Roles Suite")
}

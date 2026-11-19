package configfile_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigFile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config File Suite")
}

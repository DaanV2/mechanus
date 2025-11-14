package tracing_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/tracing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("GetConfig", func() {
		It("should return default configuration", func() {
			cfg := tracing.GetConfig()
			Expect(cfg.Enabled).To(BeFalse())
			Expect(cfg.Endpoint).To(Equal("localhost:4318"))
			Expect(cfg.ServiceName).To(Equal("mechanus-server"))
			Expect(cfg.Insecure).To(BeTrue())
		})
	})
})

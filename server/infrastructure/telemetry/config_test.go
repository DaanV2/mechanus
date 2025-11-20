package telemetry_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/telemetry"
	"github.com/DaanV2/mechanus/server/mechanus"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("GetConfig", func() {
		It("should return default configuration", func() {
			cfg := telemetry.GetConfig()
			Expect(cfg.Enabled).To(BeFalse())
			Expect(cfg.Endpoint).To(Equal("localhost:4318"))
			Expect(cfg.ServiceName).To(Equal(mechanus.SERVICE_NAME))
			Expect(cfg.Insecure).To(BeTrue())
		})
	})
})

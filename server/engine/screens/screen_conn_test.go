package screens_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/screens"
)

var _ = Describe("ConnectionInfo.CanAccess", func() {
	DescribeTable("decides whenever a connection may open a screen",
		func(info screens.ConnectionInfo, screenID string, expected bool) {
			Expect(info.CanAccess(screenID)).To(Equal(expected))
		},
		Entry("a user opens any screen", screens.ConnectionInfo{Type: devices.DeviceTypeUser}, "tv-1", true),
		Entry("a device opens its own screen", screens.ConnectionInfo{Type: devices.DeviceTypeDevice, DeviceID: "tv-1"}, "tv-1", true),
		Entry("a device can't open another screen", screens.ConnectionInfo{Type: devices.DeviceTypeDevice, DeviceID: "tv-1"}, "tv-2", false),
		Entry("a device without device ID can't open a screen", screens.ConnectionInfo{Type: devices.DeviceTypeDevice}, "tv-1", false),
		Entry("an unknown connection can't open a screen", screens.ConnectionInfo{}, "tv-1", false),
	)
})

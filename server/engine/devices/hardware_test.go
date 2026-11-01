package devices_test

import (
	"github.com/DaanV2/mechanus/server/engine/devices"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("ParseHardwareType",
	func(value string, expected devices.HardwareType) {
		Expect(devices.ParseHardwareType(value)).To(Equal(expected))
	},
	Entry("screen", "screen", devices.HardwareScreen),
	Entry("table", "table", devices.HardwareTable),
	Entry("tablet", "tablet", devices.HardwareTablet),
	Entry("phone", "phone", devices.HardwarePhone),
	Entry("desktop", "desktop", devices.HardwareDesktop),
	Entry("other", "other", devices.HardwareOther),
	Entry("case and whitespace insensitive", " Screen ", devices.HardwareScreen),
	Entry("unknown values become other", "toaster", devices.HardwareOther),
	Entry("empty becomes other", "", devices.HardwareOther),
)

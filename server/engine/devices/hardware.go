package devices

import "strings"

// HardwareType is what kind of hardware a connecting client runs on, as reported by the client in
// the X-Device-Type header. It is metadata only and never used for authorization.
type HardwareType string

// Hardware type constants.
const (
	HardwareScreen  HardwareType = "screen"  // HardwareScreen is a wall mounted or standalone screen / TV.
	HardwareTable   HardwareType = "table"   // HardwareTable is a screen lying flat in the table.
	HardwareTablet  HardwareType = "tablet"  // HardwareTablet is a tablet.
	HardwarePhone   HardwareType = "phone"   // HardwarePhone is a phone.
	HardwareDesktop HardwareType = "desktop" // HardwareDesktop is a desktop or laptop browser.
	HardwareOther   HardwareType = "other"   // HardwareOther is anything else, or an unknown value.
)

// ParseHardwareType returns the hardware type for a reported value. Unknown or empty values are
// stored as [HardwareOther], never rejected.
func ParseHardwareType(value string) HardwareType {
	switch t := HardwareType(strings.ToLower(strings.TrimSpace(value))); t {
	case HardwareScreen, HardwareTable, HardwareTablet, HardwarePhone, HardwareDesktop, HardwareOther:
		return t
	default:
		return HardwareOther
	}
}

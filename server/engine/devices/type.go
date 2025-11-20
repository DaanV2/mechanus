package devices

// DeviceType represents the type of device.
type DeviceType int

// Device type constants.
const (
	DeviceTypeUnknown DeviceType = iota // DeviceTypeUnknown represents an unknown device type.
	DeviceTypeUser                       // DeviceTypeUser represents a user device.
	DeviceTypeDevice                     // DeviceTypeDevice represents a device.
)

func (d DeviceType) String() string {
	switch d {
	case DeviceTypeUser:
		return "user"
	case DeviceTypeDevice:
		return "device"
	case DeviceTypeUnknown:
	default:
	}

	return "unknown"
}

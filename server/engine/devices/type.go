package devices

// DeviceType represents the type of device.
type DeviceType int

const (
	// DeviceTypeUnknown represents an unknown device type.
	DeviceTypeUnknown DeviceType = iota
	DeviceTypeUser
	DeviceTypeDevice
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

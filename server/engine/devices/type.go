package devices

type DeviceType int

const (
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

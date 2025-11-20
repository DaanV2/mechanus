package roles

const (
	// Admin is the administrator role with full permissions.
	Admin Role = "admin"
	Operator Role = "operator"
	User     Role = "user"
	Viewer   Role = "viewer"
	Device   Role = "device"
)

// Role represents a user role in the system.
type Role string

func (r Role) String() string {
	return string(r)
}

// Inherits checks if this role inherits the permissions of another role.
func (r Role) Inherits(other Role) bool {
	return r.value() >= other.value()
}

func (r Role) value() int {
	switch r {
	case Admin:
		return 3
	case Operator:
		return 2
	case User:
		return 1
	case Viewer, Device:
		return 0
	}

	return -1
}

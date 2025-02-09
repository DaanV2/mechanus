package roles

type Role string

const (
	DEVICE   Role = "device"   // Devices dont have permission like users, but have their own set of screens
	USER     Role = "user"     // User can only access players places, such as their characters, or screen related to characters
	OPERATOR Role = "operator" // Admin inherits all of device or user
	ADMIN    Role = "admin"    // Admin inherits all of operator, and has the most access
)

var levels = map[Role]int{
	USER:     0,
	DEVICE:   0,
	OPERATOR: 1,
	ADMIN:    2,
}

func (r Role) Inherits(role Role) bool {
	return r.Level() <= role.Level()
}

func (r Role) Level() int {
	l, ok := levels[r]
	if !ok {
		return -1
	}

	return l
}

func (r Role) String() string {
	return string(r)
}

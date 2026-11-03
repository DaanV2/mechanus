package authentication

import (
	"errors"

	"github.com/DaanV2/mechanus/server/engine/authentication/roles"
)

var (
	// ErrIdentityRequired is returned when a route needs an identity and the connection has none: answer 401.
	ErrIdentityRequired = errors.New("authentication required")
	// ErrNotAllowed is returned when the identity is known but not allowed on the route: answer 403.
	ErrNotAllowed = errors.New("not allowed")
)

// Policy says which identities may use a route. Every route registers one alongside its handler:
// authentication is optional when connecting, and enforced by the route.
type Policy struct {
	name   string
	device bool       // Paired devices are allowed
	user   bool       // Users are allowed
	role   roles.Role // Users need this role, when set
}

var (
	// PolicyAnonymous accepts any identity, including none at all.
	PolicyAnonymous = Policy{name: "anonymous"}
	// PolicyDevice accepts paired devices and users.
	PolicyDevice = Policy{name: "device", device: true, user: true}
	// PolicyUser only accepts users.
	PolicyUser = Policy{name: "user", user: true}
)

// PolicyRole only accepts users holding the role, or a role that inherits it.
func PolicyRole(role roles.Role) Policy {
	return Policy{name: "role:" + role.String(), user: true, role: role}
}

// String returns the name of the policy: anonymous, device, user or role:<role>.
func (p Policy) String() string {
	return p.name
}

// Authorize returns nil when the identity may use the route, [ErrIdentityRequired] when it needs an identity
// but there is none, and [ErrNotAllowed] when the identity isn't allowed.
func (p Policy) Authorize(identity *Identity) error {
	if !p.device && !p.user {
		return nil
	}

	switch identity.Kind {
	case IdentityDevice:
		if p.device {
			return nil
		}
	case IdentityUser:
		if p.user && (p.role == "" || roles.GrantsHasRole(identity.Claims, p.role)) {
			return nil
		}
	case IdentityAnonymous:
		return ErrIdentityRequired
	}

	return ErrNotAllowed
}

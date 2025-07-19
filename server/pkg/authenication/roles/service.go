package roles

import (
	"errors"
	"slices"
	"strings"
)

const (
	Admin    Role = "admin"
	Operator Role = "operator"
	User     Role = "user"
	Viewer   Role = "viewer"
)

type Role string

func (r Role) String() string {
	return string(r)
}

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
	case Viewer:
		return 0
	}

	return -1
}

type RoleContainer interface {
	GetRoles() []string
	SetRoles(roles ...string)
}

type RoleService struct {
}

func ParseRole(role string) (Role, error) {
	v := Role(strings.ToLower(role))
	switch v {
	case Admin, Operator, User, Viewer:
		return v, nil
	default:
		return Viewer, errors.New("unknown role: " + role)
	}
}

func (r *RoleService) GrantRole(grants RoleContainer, role Role) {
	roles := slices.Clone(grants.GetRoles())
	if slices.Contains(roles, role.String()) {
		return // already granted
	}
	roles = append(roles, role.String())
	grants.SetRoles(roles...)
}

// HasRole checks if the grants contains a role that has the given role, or inherits it.
func (r *RoleService) HasRole(grants RoleContainer, role Role) bool {
	return GrantsHasRole(grants, role)
}

func GrantsHasRole(grants RoleContainer, role Role) bool {
	for _, r := range grants.GetRoles() {
		if Role(r).Inherits(role) {
			return true
		}
	}

	return false
}
